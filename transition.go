package main

import "fmt"

type Signal struct {
	Tau float64 `json:"tau"`
	C   float64 `json:"c"`
	R   float64 `json:"r"`
	E   bool    `json:"e"`
}

type TransitionRecord struct {
	Kind   string `json:"kind"`
	Cost   int    `json:"cost,omitempty"`
	Signal Signal `json:"signal,omitempty"`
}

type TransitionOp interface {
	Name() string
	Record() TransitionRecord
	Next(State) (State, error)
}

type Advance struct {
	Signal Signal
	Cost   int
}

func (a Advance) Name() string { return "advance" }

func (a Advance) Record() TransitionRecord {
	return TransitionRecord{
		Kind:   a.Name(),
		Cost:   a.Cost,
		Signal: a.Signal,
	}
}

func (a Advance) Next(prev State) (State, error) {
	if a.Cost < 0 {
		return State{}, fmt.Errorf("cost must be non-negative")
	}
	if a.Cost > prev.Budget {
		return State{}, fmt.Errorf("budget violation: cost %d exceeds budget %d", a.Cost, prev.Budget)
	}

	next := prev
	next.Phase = candidatePhase(a.Signal)
	next.Budget = prev.Budget - a.Cost
	next.Sequence++
	return next, nil
}

type Commit struct{}

func (Commit) Name() string { return "commit" }

func (Commit) Record() TransitionRecord {
	return TransitionRecord{Kind: "commit"}
}

func (Commit) Next(prev State) (State, error) {
	next := prev
	next.Committed = true
	next.Sequence++
	return next, nil
}

func candidatePhase(signal Signal) Phase {
	if signal.E || signal.R >= 0.85 {
		return Escalating
	}
	if signal.C >= 0.70 {
		return Constrained
	}
	if signal.Tau >= 0.40 {
		return Transition
	}
	return Stable
}

func opFromRecord(record TransitionRecord) (TransitionOp, error) {
	switch record.Kind {
	case "advance":
		return Advance{Signal: record.Signal, Cost: record.Cost}, nil
	case "commit":
		return Commit{}, nil
	default:
		return nil, fmt.Errorf("unknown transition kind: %q", record.Kind)
	}
}

func Apply(prev State, op TransitionOp) (State, error) {
	if err := prev.Validate(); err != nil {
		return State{}, err
	}
	next, err := op.Next(prev)
	if err != nil {
		return State{}, err
	}
	if err := enforceInvariants(prev, next, op.Name()); err != nil {
		return State{}, err
	}
	return next, nil
}
