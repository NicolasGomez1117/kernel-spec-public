package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeterministicReplayEquivalence(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	initial := InitialState(10)

	state := initial
	var err error
	state, err = ApplyToJournal(path, state, Advance{
		Signal: Signal{Tau: 0.5},
		Cost:   2,
	})
	if err != nil {
		t.Fatal(err)
	}
	state, err = ApplyToJournal(path, state, Advance{
		Signal: Signal{C: 0.8},
		Cost:   3,
	})
	if err != nil {
		t.Fatal(err)
	}
	state, err = ApplyToJournal(path, state, Commit{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Apply(state, Commit{}); err == nil || !strings.Contains(err.Error(), "commit is absorbing") {
		t.Fatalf("expected absorbing commit, got %v", err)
	}

	replayed, err := Replay(path, initial)
	if err != nil {
		t.Fatal(err)
	}
	if replayed != state {
		t.Fatalf("replayed state %+v, want %+v", replayed, state)
	}
}

func TestBudgetViolationDetection(t *testing.T) {
	_, err := Apply(InitialState(1), Advance{
		Signal: Signal{Tau: 0.5},
		Cost:   2,
	})
	if err == nil || !strings.Contains(err.Error(), "budget violation") {
		t.Fatalf("expected budget violation, got %v", err)
	}
}

func TestIllegalTransitionRejection(t *testing.T) {
	prev := State{
		Phase:    Escalating,
		Budget:   4,
		Sequence: 2,
	}

	_, err := Apply(prev, Advance{
		Signal: Signal{},
		Cost:   1,
	})
	if err == nil || !strings.Contains(err.Error(), "illegal transition") {
		t.Fatalf("expected illegal transition, got %v", err)
	}
}

func TestMutationBasedReplayFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "journal.jsonl")
	initial := InitialState(5)

	state, err := ApplyToJournal(path, initial, Advance{
		Signal: Signal{Tau: 0.5},
		Cost:   1,
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = ApplyToJournal(path, state, Commit{})
	if err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(content), `"budget":4`, `"budget":3`, 1)
	if mutated == string(content) {
		t.Fatal("failed to mutate journal fixture")
	}
	if err := os.WriteFile(path, []byte(mutated), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = Replay(path, initial)
	if err == nil || !strings.Contains(err.Error(), "hash mismatch") {
		t.Fatalf("expected hash mismatch, got %v", err)
	}
}
