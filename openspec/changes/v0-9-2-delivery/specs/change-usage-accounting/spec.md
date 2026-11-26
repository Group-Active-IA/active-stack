## ADDED Requirements

### Requirement: Auditable usage through PR closure

Delivery SHALL record active work intervals, exact model identifiers, measured tokens by billable category, API rates in USD per million and their sources for the branch. Cumulative counters and overlapping work intervals SHALL be normalized before aggregation.

#### Scenario: Multiple models or rate tiers contribute

- **WHEN** the branch uses different models or billing categories
- **THEN** the archive retains each model and rate category, computes each USD subtotal with decimal arithmetic, and records the summed active time and cost

#### Scenario: Evidence is missing

- **WHEN** required time, usage, rate or PR closure evidence cannot be obtained
- **THEN** delivery identifies what is missing and leaves archive completion pending
