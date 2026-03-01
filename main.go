package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fail("usage: admissible-transition-lab <apply|replay> [flags]")
	}

	switch os.Args[1] {
	case "apply":
		runApply(os.Args[2:])
	case "replay":
		runReplay(os.Args[2:])
	default:
		fail("unknown command")
	}
}

func runApply(args []string) {
	fs := flag.NewFlagSet("apply", flag.ExitOnError)
	journal := fs.String("journal", "journal.jsonl", "path to journal")
	budget := fs.Int("budget", 10, "initial budget when journal is empty")
	cost := fs.Int("cost", 0, "budget cost for advance")
	tau := fs.Float64("tau", 0, "transition pressure")
	c := fs.Float64("c", 0, "constraint pressure")
	r := fs.Float64("r", 0, "escalation pressure")
	e := fs.Bool("e", false, "hard escalation trigger")
	commit := fs.Bool("commit", false, "apply commit instead of advance")
	fs.Parse(args)

	state := InitialState(*budget)
	if last, err := Replay(*journal, state); err == nil {
		state = last
	} else if !os.IsNotExist(err) {
		fail(err.Error())
	}

	var op TransitionOp = Advance{
		Signal: Signal{Tau: *tau, C: *c, R: *r, E: *e},
		Cost:   *cost,
	}
	if *commit {
		op = Commit{}
	}

	next, err := ApplyToJournal(*journal, state, op)
	if err != nil {
		fail(err.Error())
	}
	emit(next)
}

func runReplay(args []string) {
	fs := flag.NewFlagSet("replay", flag.ExitOnError)
	journal := fs.String("journal", "journal.jsonl", "path to journal")
	budget := fs.Int("budget", 10, "initial budget")
	fs.Parse(args)

	state, err := Replay(*journal, InitialState(*budget))
	if err != nil {
		fail(err.Error())
	}
	emit(state)
}

func emit(v any) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	fmt.Println(string(data))
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
