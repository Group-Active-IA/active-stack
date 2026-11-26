"""Calculate and validate the time and metered token cost for a merged change."""

import argparse
from collections import defaultdict
from datetime import date, datetime, timezone
from decimal import Decimal, localcontext
import json
import re
import sys
from urllib.parse import urlsplit

from team import CREDENTIAL_RE


class EvidenceError(ValueError):
    pass


def record(value, required, path, optional=()):
    if not isinstance(value, dict):
        raise EvidenceError(f"{path}: expected an object")
    expected = set(required)
    missing = expected - value.keys()
    if missing:
        raise EvidenceError(f"{path}: missing {', '.join(sorted(missing))}")
    if value.keys() - expected - set(optional):
        raise EvidenceError(f"{path}: unexpected fields outside the evidence schema")
    return value


def entries(value, path):
    if not isinstance(value, list):
        raise EvidenceError(f"{path}: expected a list")
    return value


def safe_text(value, path, pattern=r"[^\r\n\x00-\x1f\x7f]{1,300}"):
    if (not isinstance(value, str) or not value.strip() or
            not re.fullmatch(pattern, value) or CREDENTIAL_RE.search(value)):
        raise EvidenceError(f"{path}: expected non-sensitive text in the required format")
    return value


def public_url(value, path, url_path=None):
    safe_text(value, path)
    parsed = urlsplit(value)
    if (parsed.scheme != "https" or not parsed.hostname or parsed.username is not None
            or parsed.query or any(character.isspace() for character in value)):
        raise EvidenceError(f"{path}: use an HTTPS URL without credentials or query parameters")
    if url_path and (parsed.fragment or not re.fullmatch(url_path, parsed.path)):
        raise EvidenceError(f"{path}: expected a GitHub PR URL")
    return value


def instant(value, path):
    safe_text(value, path, r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d{1,6})?(?:Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)")
    try:
        value = value[:-1] + "+00:00" if value.endswith("Z") else value
        return datetime.fromisoformat(value).astimezone(timezone.utc)
    except (ValueError, OverflowError) as exc:
        raise EvidenceError(f"{path}: invalid timestamp") from exc


def duration(start, end):
    elapsed = end - start
    return Decimal(elapsed.days * 86400 + elapsed.seconds) + Decimal(elapsed.microseconds).scaleb(-6)


def active_time(intervals):
    active_start, active_end = intervals[0]
    total = Decimal(0)
    for start, end in intervals[1:]:
        if start > active_end:
            total += duration(active_start, active_end)
            active_start, active_end = start, end
        else:
            active_end = max(active_end, end)
    return total + duration(active_start, active_end)


def integer_count(value, path):
    if type(value) is not int or value < 0:
        raise EvidenceError(f"{path}: expected a measured nonnegative integer token count")
    return value


def pr_evidence(pr, branch):
    record(pr, ("state", "headRefName", "baseRefName", "mergedAt", "mergeCommit", "url"), "pr")
    if pr["state"] != "MERGED" or pr["headRefName"] != branch:
        raise EvidenceError("pr: expected GitHub to confirm a merge of this branch")
    safe_text(pr["baseRefName"], "pr.baseRefName")
    public_url(pr["url"], "pr.url", r"/[\w.-]+/[\w.-]+/pull/[1-9][0-9]*")
    commit = record(pr["mergeCommit"], ("oid",), "pr.mergeCommit")
    safe_text(commit["oid"], "pr.mergeCommit.oid", r"(?:[0-9a-f]{40}|[0-9a-f]{64})")
    return instant(pr["mergedAt"], "pr.mergedAt")


def collect_intervals(raw, merged_at):
    spans = set()
    for index, item in enumerate(entries(raw, "intervals")):
        path = f"intervals[{index}]"
        record(item, ("start", "end", "source"), path)
        start = instant(item["start"], f"{path}.start")
        end = instant(item["end"], f"{path}.end")
        source = safe_text(item["source"], f"{path}.source")
        if not start < end <= merged_at:
            raise EvidenceError(f"{path}: expected positive elapsed time ending on or before merge")
        spans.add((start, end, source))
    if not spans:
        raise EvidenceError("intervals: record at least one measured interval")
    return sorted(spans)


def collect_rates(raw):
    rates = {}
    for index, rate in enumerate(entries(raw, "rates")):
        path = f"rates[{index}]"
        record(rate, ("id", "provider", "model", "tier", "category", "usd_per_million", "source", "checked_on"), path)
        for field in ("id", "provider", "model", "tier", "category"):
            safe_text(rate[field], f"{path}.{field}")
        public_url(rate["source"], f"{path}.source")
        checked_on = safe_text(rate["checked_on"], f"{path}.checked_on", r"\d{4}-\d{2}-\d{2}")
        try:
            date.fromisoformat(checked_on)
        except ValueError as exc:
            raise EvidenceError(f"{path}.checked_on: invalid date") from exc
        if rate["category"].casefold() in {"all", "total", "total_tokens", "unknown"}:
            raise EvidenceError(f"{path}.category: use the provider's billable category")
        safe_text(rate["usd_per_million"], f"{path}.usd_per_million", r"[0-9]{1,10}(?:\.[0-9]{1,12})?")
        if rate["id"] in rates:
            raise EvidenceError(f"{path}.id: rate identifier already defined")
        rates[rate["id"]] = rate
    return rates


def unique_usage(raw):
    seen = {}
    for index, item in enumerate(entries(raw, "usage")):
        path = f"usage[{index}]"
        record(item, ("id", "source", "billing", "rate_id"), path,
               ("tokens", "counter_id", "start", "end", "before", "after"))
        key = safe_text(item["id"], f"{path}.id")
        if key in seen:
            if item != seen[key][1]:
                raise EvidenceError(f"{path}.id: conflicting measurements for the same identifier")
        else:
            seen[key] = (path, item)
    return seen.values()


def collect_usage(raw, rates, first_start, merged_at):
    totals = defaultdict(int)
    counters = defaultdict(list)
    measurements = []
    counter_fields = {"counter_id", "start", "end", "before", "after"}
    for path, item in unique_usage(raw):
        safe_text(item["source"], f"{path}.source")
        rate_id = safe_text(item["rate_id"], f"{path}.rate_id")
        if rate_id not in rates:
            raise EvidenceError(f"{path}.rate_id: missing a verified rate for this measurement")
        billing = item["billing"]
        if billing not in ("api", "subscription"):
            raise EvidenceError(f"{path}.billing: use api or subscription")
        supplied_counters = counter_fields.intersection(item)
        if "tokens" in item and not supplied_counters:
            tokens = integer_count(item["tokens"], f"{path}.tokens")
        elif "tokens" not in item and supplied_counters == counter_fields:
            counter = safe_text(item["counter_id"], f"{path}.counter_id")
            start = instant(item["start"], f"{path}.start")
            end = instant(item["end"], f"{path}.end")
            before = integer_count(item["before"], f"{path}.before")
            after = integer_count(item["after"], f"{path}.after")
            if not first_start <= start < end <= merged_at or after < before:
                raise EvidenceError(f"{path}: invalid cumulative counter span or decreasing count")
            counters[counter].append((start, end, before, after, path))
            tokens = after - before
        else:
            raise EvidenceError(f"{path}: use a measured delta or one complete cumulative counter pair")
        totals[(billing, rate_id)] += tokens
        measurements.append(dict(item, measured_tokens=tokens))

    for windows in counters.values():
        windows.sort()
        for previous, current in zip(windows, windows[1:]):
            previous_end, previous_after = previous[1], previous[3]
            start, _, before, _, path = current
            if start < previous_end or before < previous_after:
                raise EvidenceError(f"{path}: cumulative counter windows overlap or decrease")
            if start == previous_end and before != previous_after:
                raise EvidenceError(f"{path}: adjacent readings of the same counter disagree")
    return totals, measurements


def calculate(evidence):
    record(evidence, ("branch", "pr", "intervals", "usage_status", "rates", "usage"), "evidence")
    branch = safe_text(evidence["branch"], "branch")
    merged_at = pr_evidence(evidence["pr"], branch)
    intervals = collect_intervals(evidence["intervals"], merged_at)
    first_start = intervals[0][0]
    rates = collect_rates(evidence["rates"])
    totals, measurements = collect_usage(evidence["usage"], rates, first_start, merged_at)

    status = evidence["usage_status"]
    if status == "complete" and not measurements:
        raise EvidenceError("usage: measured per-model usage is required")
    if status == "none" and (measurements or rates):
        raise EvidenceError("usage_status: none is reserved for changes with zero AI usage")
    if status not in ("complete", "none"):
        raise EvidenceError("usage_status: incomplete evidence blocks archive")

    api = Decimal(0)
    subscription = Decimal(0)
    rows = []
    with localcontext() as context:
        max_digits = max((len(str(count)) for count in totals.values()), default=1)
        context.prec = max(28, max_digits + 24 + len(str(len(totals))))
        for (billing, rate_id), tokens in sorted(totals.items()):
            rate = rates[rate_id]
            cost = (tokens * Decimal(rate["usd_per_million"])).scaleb(-6)
            if billing == "api":
                api += cost
            else:
                subscription += cost
            rows.append(dict(rate, billing=billing, tokens=tokens, token_cost_usd=format(cost, "f")))
        total_usd = api + subscription

    return {
        "ok": True,
        "branch": branch,
        "pr": evidence["pr"],
        "intervals": [{"start": start.isoformat(), "end": end.isoformat(), "source": source}
                      for start, end, source in intervals],
        "active_seconds": format(active_time(sorted({(start, end) for start, end, _ in intervals})), "f"),
        "elapsed_to_merge_seconds": format(duration(first_start, merged_at), "f"),
        "usage_status": status,
        "measurements": measurements,
        "model_rates": rows,
        "total_tokens": sum(totals.values()),
        "api_rate_cost_usd": format(api, "f"),
        "subscription_api_equivalent_usd": format(subscription, "f"),
        "total_at_api_rates_usd": format(total_usd, "f"),
    }


def unique_object(pairs):
    document = {}
    for key, value in pairs:
        if key in document:
            raise EvidenceError("JSON contains a duplicate field")
        document[key] = value
    return document


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("evidence", help="path to the sanitized usage evidence JSON")
    args = parser.parse_args()
    try:
        with open(args.evidence, encoding="utf-8-sig") as source:
            report = calculate(json.load(source, object_pairs_hook=unique_object))
    except (OSError, ValueError) as exc:
        print(json.dumps({"ok": False, "error": str(exc)}, ensure_ascii=False))
        return 1
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 0


if __name__ == "__main__":
    sys.exit(main())
