## Direct or delegated

Decide per action, in this order. The first step that applies closes the decision.

1. **Independent review required** (for example, verify after an apply): the review goes to a
   different agent than the one that built it. Nothing below overrides this. If delegation is not
   available, stop that verification, say what is missing, and continue with the rest.
2. **File conflict**: the write set overlaps an open assignment → serialize or narrow the scope.
3. **Work directly** when the task is short, depends on decisions made in this conversation, or you
   will read the full result anyway.
4. **Delegate** when the task produces more than 300 lines or more than 10,000 characters of
   output, when the answer is a conclusion and not the content read, or when two subtasks can run
   in parallel.
5. Otherwise, work directly.

Record each decision in the round report as `{action, step, signal, direct|delegated|stopped}`.
When you delegate, read `modules/delegation.md` first.
