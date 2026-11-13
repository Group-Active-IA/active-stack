# Delegation and worker return

Choose delegation using the selective-delegation policy of change 07. Prefer direct execution unless the task or a mandatory independent review requires delegation. Missing optional delegation does not stop unrelated actions. When independent verification is required, use a reviewer distinct from the constructor or stop that verification.

Each delegated assignment identifies the task, base commit and tested tree state, allowed files, acceptance criteria, applicable universal rules and action modules. Allow at most two writing workers in parallel, each in its own worktree, with disjoint independent scopes. Serialize overlapping or dependent work. Validate changes against each worker's recorded starting state and verify the combined result before integration.

Use the return envelope owned by change 07, with the receipts field owned by change 06: estado, partida, modificados, hallazgos, recibos, pendiente. State which revision of those contracts was supplied. A self-reported result remains unverified until execution and tested code state are independently corroborated and the evidence covers the requirement. The same evidence standard applies to direct execution.

This module does not provide the capture mechanism or corrected receipt schema of changes 05 and 06. If either required dependency is unavailable, stop the affected evidence action and report the missing contract or protection.
