package filemerge

import (
	"sort"
	"strings"
)

// markerPrefix/closePrefix are the CURRENT (canonical) marker delimiters —
// the only ones ever used when WRITING a section (openMarker/closeMarker,
// and the append path in InjectMarkdownSection).
const (
	markerPrefix = "<!-- active-stack:"
	markerSuffix = " -->"
	closePrefix  = "<!-- /active-stack:"
)

// legacyMarkerPrefixPairs lists marker prefix pairs from previous installer
// names that MarkedSectionIDs/InjectMarkdownSection must still recognize
// when LOCATING an existing section. This project was previously named
// "JR Stack" and wrote markers as "<!-- jr-stack:ID -->" before the rebrand
// to "active-stack:" (commit 2f8336e); no migration was ever added, so a
// user upgrading from that version has orphaned jr-stack: blocks in their
// CLAUDE.md/AGENTS.md that neither re-install nor uninstall ever touched.
// Recognizing the legacy prefix here — for LOOKUP only, never for writing —
// lets the existing purge/replace machinery (PurgeStaleSections, Inject)
// clean those up for free, the same way it already handles any other
// stale/renamed section.
var legacyMarkerPrefixPairs = []markerPair{
	{open: "<!-- jr-stack:", close: "<!-- /jr-stack:"},
}

type markerPair struct{ open, close string }

// recognizedMarkerPairs returns every prefix pair MarkedSectionIDs/
// InjectMarkdownSection should recognize when locating a section, current
// prefix first (so it always wins when both a current and a legacy block
// for the same ID exist).
func recognizedMarkerPairs() []markerPair {
	pairs := make([]markerPair, 0, 1+len(legacyMarkerPrefixPairs))
	pairs = append(pairs, markerPair{open: markerPrefix, close: closePrefix})
	pairs = append(pairs, legacyMarkerPrefixPairs...)
	return pairs
}

// sectionOccurrence is one well-formed section found while scanning content
// under a single marker prefix pair, keeping its position for document-order
// merging across prefixes.
type sectionOccurrence struct {
	id  string
	pos int
}

// MarkedSectionIDs returns the distinct section IDs that have a well-formed
// marker pair (the current active-stack: prefix, or a recognized legacy
// prefix — see legacyMarkerPrefixPairs) in content, in first-seen document
// order. When the same ID appears under more than one prefix, it is reported
// once (first occurrence wins).
//
// Malformed markers are ignored — an opening marker with no matching close, an
// ID containing whitespace, or a stray closing marker are never reported. This
// keeps callers that purge by ID from ever acting on a half-written section.
func MarkedSectionIDs(content string) []string {
	var occurrences []sectionOccurrence
	for _, pair := range recognizedMarkerPairs() {
		occurrences = append(occurrences, scanSectionOccurrences(content, pair)...)
	}
	sort.SliceStable(occurrences, func(i, j int) bool { return occurrences[i].pos < occurrences[j].pos })

	seen := make(map[string]bool)
	var ids []string
	for _, occ := range occurrences {
		if seen[occ.id] {
			continue
		}
		seen[occ.id] = true
		ids = append(ids, occ.id)
	}
	return ids
}

// scanSectionOccurrences finds every well-formed section under a single
// marker prefix pair, recording each open marker's position for document-
// order merging in MarkedSectionIDs.
func scanSectionOccurrences(content string, pair markerPair) []sectionOccurrence {
	var out []sectionOccurrence

	offset := 0
	for {
		rel := strings.Index(content[offset:], pair.open)
		if rel < 0 {
			break
		}
		pos := offset + rel
		idStart := pos + len(pair.open)
		end := strings.Index(content[idStart:], markerSuffix)
		if end < 0 {
			break
		}
		id := content[idStart : idStart+end]
		offset = idStart + end + len(markerSuffix)

		if id == "" || strings.ContainsAny(id, " \t\r\n") {
			continue
		}
		if strings.Contains(content, pair.close+id+markerSuffix) {
			out = append(out, sectionOccurrence{id: id, pos: pos})
		}
	}

	return out
}

// openMarker returns the CURRENT opening marker for a section ID (writing only).
func openMarker(sectionID string) string {
	return markerPrefix + sectionID + markerSuffix
}

// closeMarker returns the CURRENT closing marker for a section ID (writing only).
func closeMarker(sectionID string) string {
	return closePrefix + sectionID + markerSuffix
}

// locateSection finds the open/close marker span for sectionID, trying each
// recognized prefix in order (current first, then legacy) so a section
// written under a previous installer name is still found for replacement or
// removal. ok is false when no well-formed pair is found under any prefix.
func locateSection(existing, sectionID string) (openIdx, closeIdx int, closeMarkerLen int, ok bool) {
	for _, pair := range recognizedMarkerPairs() {
		om := pair.open + sectionID + markerSuffix
		cm := pair.close + sectionID + markerSuffix

		oi := strings.Index(existing, om)
		ci := strings.Index(existing, cm)
		if oi >= 0 && ci >= 0 && ci > oi {
			return oi, ci, len(cm), true
		}
	}
	return 0, 0, 0, false
}

// InjectMarkdownSection replaces or appends a marked section in a markdown file.
// Markers use HTML comments: <!-- active-stack:SECTION_ID --> ... <!-- /active-stack:SECTION_ID -->
// An existing section is located under any recognized prefix (current or
// legacy — see legacyMarkerPrefixPairs), but is always REPLACED using the
// current prefix, so a legacy-prefixed section is transparently upgraded the
// next time it is injected. If the section doesn't exist under any
// recognized prefix, it's appended at the end using the current prefix.
// Content outside markers is never touched.
// If content is empty, the section (including markers) is removed.
func InjectMarkdownSection(existing, sectionID, content string) string {
	openIdx, closeIdx, closeLen, found := locateSection(existing, sectionID)

	// If both markers are found and in the correct order, replace the section.
	if found {
		// If content is empty, remove the entire section including markers.
		if content == "" {
			before := existing[:openIdx]
			after := existing[closeIdx+closeLen:]

			// Clean up trailing newline after close marker.
			if len(after) > 0 && after[0] == '\n' {
				after = after[1:]
			}
			// Clean up trailing newline before open marker.
			result := strings.TrimRight(before, "\n")
			if after != "" {
				if result != "" {
					result += "\n"
				}
				result += after
			} else if result != "" {
				result += "\n"
			}
			return result
		}

		before := existing[:openIdx]
		after := existing[closeIdx+closeLen:]

		var sb strings.Builder
		sb.WriteString(before)
		sb.WriteString(openMarker(sectionID))
		sb.WriteString("\n")
		sb.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString(closeMarker(sectionID))
		sb.WriteString(after)
		return sb.String()
	}

	// If content is empty and section doesn't exist, return existing unchanged.
	if content == "" {
		return existing
	}

	// Section not found — append at end.
	var sb strings.Builder
	sb.WriteString(existing)
	if existing != "" && !strings.HasSuffix(existing, "\n") {
		sb.WriteString("\n")
	}
	if existing != "" {
		sb.WriteString("\n")
	}
	sb.WriteString(openMarker(sectionID))
	sb.WriteString("\n")
	sb.WriteString(content)
	if !strings.HasSuffix(content, "\n") {
		sb.WriteString("\n")
	}
	sb.WriteString(closeMarker(sectionID))
	sb.WriteString("\n")
	return sb.String()
}
