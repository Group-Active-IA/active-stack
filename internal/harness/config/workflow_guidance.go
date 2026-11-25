package config

const workflowGuidance = "Handle routine tasks directly. Recommend an OPSX proposal only when proposal, specs, design, and tasks help agree on and verify a complex change. For tool-specific work, point to the installed skill or client command."

const codexDeliveryGuidance = "When installed, invoke `$active-delivery` to agree on the delivery workflow for a task. Phases of the project foundation run by `active-orchestrator` are not delivery units: do not ask for ODD or SDD for them."

const claudeDeliveryGuidance = "Before starting or resuming a unit of project work, such as a roadmap item, a feature or a fix, invoke the `active-delivery` skill when it is installed and let it agree on ODD or SDD with the user. The guidance below on routine tasks applies only when that skill is unavailable. Phases of the project foundation run by `active-orchestrator` are not delivery units: do not ask for ODD or SDD for them."

// deliveryAccounting applies to both the coordinator and direct OpenSpec archive actions.
const deliveryAccounting = "For a delivery tracked by Active Delivery, start branch time and token accounting with the first implementation session. Before archiving the change, including direct OpenSpec archive requests, follow the active-delivery usage accounting reference and validate the evidence for the merged PR. Archive remains pending when required evidence is missing."
