package main

import "fmt"

func ApplyToJournal(path string, current State, op TransitionOp) (State, error) {
	next, err := Apply(current, op)
	if err != nil {
		return State{}, err
	}
	if err := appendJournal(path, current, next, op); err != nil {
		return State{}, err
	}
	return next, nil
}

func Replay(path string, initial State) (State, error) {
	if err := verifyJournal(path); err != nil {
		return State{}, err
	}

	entries, err := readJournal(path)
	if err != nil {
		return State{}, err
	}

	state := initial
	for _, entry := range entries {
		if entry.Prev != state {
			return State{}, fmt.Errorf("non-canonical previous state at entry %d", entry.Index)
		}

		op, err := opFromRecord(entry.Op)
		if err != nil {
			return State{}, err
		}
		next, err := Apply(entry.Prev, op)
		if err != nil {
			return State{}, err
		}
		if next != entry.Next {
			return State{}, fmt.Errorf("replay divergence at entry %d", entry.Index)
		}
		state = next
	}

	return state, nil
}
