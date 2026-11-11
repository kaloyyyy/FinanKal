# Clear Installment Entry Design

The installment dialog mixes automatic equal splitting with manual row editing. Make the chosen workflow explicit so users understand what drives the schedule.

## Design

Offer two modes: **Split a total** (total amount, installment count, first due date; generates editable monthly rows) and **Enter manually** (users add each due date and amount directly, with no generator fields). Keep payment source and plan name shared. Both modes submit the current schedule payload, so engine and API behavior stay unchanged.

## Approaches

Clarify the current mixed form with labels, or separate the two workflows. Choose separate modes because manual edits should not be overwritten by generator inputs.

## Constraints

Keep existing validation, centavo distribution, monthly date generation, account selection, and schedule payload. No build, tests, lint, or preview.
