package clustershared

import "fmt"

// MapDesiredStateTransition is a pure helper that translates a
// (prev, next) `desired_state` pair into the corresponding cluster
// lifecycle operation. Callers feed the returned op (or OpNoop) into
// the same ApplyUpdates pipeline RouteUpdate uses.
//
// Returned OperationKind values:
//   - OpHibernate when transitioning running → hibernated.
//   - OpResume    when transitioning hibernated → running.
//   - OpNoop      when prev == next OR next is empty.
//
// Errors:
//   - Unknown next value (not in {"running", "hibernated"}).
//   - Cross-state transition that doesn't match the documented
//     hibernate/resume pair (e.g. running → "destroyed").
const OpNoop OperationKind = "noop"

// MapDesiredStateTransition is documented above.
func MapDesiredStateTransition(prev, next string) (OperationKind, error) {
	if next == "" || prev == next {
		return OpNoop, nil
	}
	switch next {
	case "hibernated":
		if prev != "" && prev != "running" {
			return "", fmt.Errorf("clustershared: cannot transition from %q to hibernated; cluster must be running", prev)
		}
		return OpHibernate, nil
	case "running":
		if prev != "" && prev != "hibernated" && prev != "running" {
			return "", fmt.Errorf("clustershared: cannot transition from %q to running; cluster must be hibernated", prev)
		}
		if prev == "" {
			return OpNoop, nil
		}
		return OpResume, nil
	default:
		return "", fmt.Errorf("clustershared: invalid desired_state %q (allowed: running, hibernated)", next)
	}
}
