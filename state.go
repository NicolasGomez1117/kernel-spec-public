package main

import "fmt"

type Phase string

const (
	Stable      Phase = "Stable"
	Transition  Phase = "Transition"
	Constrained Phase = "Constrained"
	Escalating  Phase = "Escalating"
)

type State struct {
	Phase     Phase `json:"phase"`
	Budget    int   `json:"budget"`
	Committed bool  `json:"committed"`
	Sequence  int   `json:"sequence"`
}

func InitialState(budget int) State {
	return State{
		Phase:  Stable,
		Budget: budget,
	}
}

func (s State) Validate() error {
	if rankOf(s.Phase) < 0 {
		return fmt.Errorf("invalid phase: %q", s.Phase)
	}
	if s.Budget < 0 {
		return fmt.Errorf("budget must be non-negative")
	}
	if s.Sequence < 0 {
		return fmt.Errorf("sequence must be non-negative")
	}
	return nil
}

func rankOf(p Phase) int {
	switch p {
	case Stable:
		return 0
	case Transition:
		return 1
	case Constrained:
		return 2
	case Escalating:
		return 3
	default:
		return -1
	}
}
