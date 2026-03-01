package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

const zeroHash = "0000000000000000000000000000000000000000000000000000000000000000"

type JournalEntry struct {
	Index    int              `json:"index"`
	PrevHash string           `json:"prev_hash"`
	Hash     string           `json:"hash"`
	Prev     State            `json:"prev"`
	Next     State            `json:"next"`
	Op       TransitionRecord `json:"op"`
}

func (e JournalEntry) payloadJSON() ([]byte, error) {
	type payload struct {
		Index    int              `json:"index"`
		PrevHash string           `json:"prev_hash"`
		Prev     State            `json:"prev"`
		Next     State            `json:"next"`
		Op       TransitionRecord `json:"op"`
	}
	return json.Marshal(payload{
		Index:    e.Index,
		PrevHash: e.PrevHash,
		Prev:     e.Prev,
		Next:     e.Next,
		Op:       e.Op,
	})
}

func entryHash(prevHash string, payload []byte) string {
	sum := sha256.Sum256(append(append([]byte(prevHash), ':'), payload...))
	return hex.EncodeToString(sum[:])
}

func appendJournal(path string, prev, next State, op TransitionOp) error {
	last, err := lastEntry(path)
	if err != nil {
		return err
	}

	entry := JournalEntry{
		Index:    prev.Sequence + 1,
		PrevHash: zeroHash,
		Prev:     prev,
		Next:     next,
		Op:       op.Record(),
	}
	if last != nil {
		entry.PrevHash = last.Hash
	}

	payload, err := entry.payloadJSON()
	if err != nil {
		return err
	}
	entry.Hash = entryHash(entry.PrevHash, payload)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	line, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	_, err = f.Write(append(line, '\n'))
	return err
}

func readJournal(path string) ([]JournalEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var entries []JournalEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var entry JournalEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return entries, nil
}

func lastEntry(path string) (*JournalEntry, error) {
	entries, err := readJournal(path)
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, nil
	}
	last := entries[len(entries)-1]
	return &last, nil
}

func verifyJournal(path string) error {
	entries, err := readJournal(path)
	if err != nil {
		return err
	}

	prevHash := zeroHash
	for i, entry := range entries {
		if entry.Index != i+1 {
			return fmt.Errorf("index mismatch at entry %d", i+1)
		}
		if entry.PrevHash != prevHash {
			return fmt.Errorf("prev hash mismatch at entry %d", entry.Index)
		}

		payload, err := entry.payloadJSON()
		if err != nil {
			return err
		}
		expected := entryHash(entry.PrevHash, payload)
		if entry.Hash != expected {
			return fmt.Errorf("hash mismatch at entry %d", entry.Index)
		}
		prevHash = entry.Hash
	}

	return nil
}
