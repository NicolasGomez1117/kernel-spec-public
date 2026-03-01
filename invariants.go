package main

import "fmt"

func legalTransition(prev, next Phase) bool {
	return rankOf(next) >= rankOf(prev) && rankOf(prev) >= 0 && rankOf(next) >= 0
}

func enforceInvariants(prev, next State, transitionName string) error {
	if err := prev.Validate(); err != nil {
		return err
	}
	if err := next.Validate(); err != nil {
		return err
	}
	if !legalTransition(prev.Phase, next.Phase) {
		return fmt.Errorf("illegal transition %s -> %s for %s", prev.Phase, next.Phase, transitionName)
	}
	if next.Budget > prev.Budget {
		return fmt.Errorf("budget increased from %d to %d", prev.Budget, next.Budget)
	}
	if prev.Committed {
		return fmt.Errorf("commit is absorbing")
	}
	if transitionName == "commit" && !next.Committed {
		return fmt.Errorf("commit must set committed")
	}
	if transitionName != "commit" && next.Committed != prev.Committed {
		return fmt.Errorf("non-commit transition changed committed")
	}
	if next.Sequence != prev.Sequence+1 {
		return fmt.Errorf("sequence must advance by one")
	}
	return nil
}
