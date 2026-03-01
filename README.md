# Admissible Transition Lab

This repository is a compact Go lab for deterministic transition semantics: a
four-state monotone lattice, an append-only JSONL journal, and replay that
re-executes the same transition logic used on the write path. `State` carries
phase, budget, commit status, and sequence. `Advance` derives a candidate
phase from `(tau, c, r, e)` using fixed thresholds. `Commit` is terminal.
Every accepted step is hash-linked in the journal.
Guards enforced on apply and replay:

- budget never increases
- commit is absorbing
- phase regressions are rejected
- replay must reconstruct the stored next state exactly
CLI:

```sh
go run . apply -journal ./journal.jsonl -budget 10 -tau 0.5 -cost 2
go run . apply -journal ./journal.jsonl -c 0.8 -cost 3
go run . apply -journal ./journal.jsonl -commit
go run . replay -journal ./journal.jsonl -budget 10
```

Tests cover deterministic replay equivalence, budget violation detection,
illegal transition rejection, and replay failure after journal mutation.
