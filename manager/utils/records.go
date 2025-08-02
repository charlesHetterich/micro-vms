package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

const (
	minID = 5
	maxID = 250
)

// Micro VM record.
type Record struct {
	ID  string `json:"id"`
	PID int    `json:"pid"`
}

// Handles management of micro-vm records stored to disk
type RecordKeeper struct {
	filePath string
	mu       sync.Mutex
}

// Structure of `RecordKeeper`'s data blob
type store = map[string]Record

// Provide 0 padding for record IDs
func padID(id string) string {
	if len(id) < 3 {
		return fmt.Sprintf("%03s", id)
	}
	return id
}

// Load store from file
func (rk *RecordKeeper) loadStore() (store, error) {
	// Create store if it does not exist
	if _, err := os.Stat(rk.filePath); errors.Is(err, os.ErrNotExist) {
		dir := filepath.Dir(rk.filePath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		return nil, os.WriteFile(rk.filePath, []byte("{}\n"), 0o644)
	}

	// Read raw store blob
	raw, err := os.ReadFile(rk.filePath)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return store{}, nil
	}

	// Parse store & return
	var s store
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", rk.filePath, err)
	}
	if s == nil {
		s = store{}
	}
	return s, nil
}

// Save store to file
// (writes to tmp and then overwrite true file to avoid corruption)
func (rk *RecordKeeper) save(s store) error {
	tmp := rk.filePath + ".tmp"
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, rk.filePath)
}

// Finds available ID for a next record to use. IDs are
//  1. unique,
//  2. in range [5, 250]
//  3. padded to 3 digits
func nextAvailableID(s store) (string, error) {
	for idRaw := minID; idRaw <= maxID; idRaw++ {
		id := strconv.Itoa(idRaw)
		if _, exists := s[padID(id)]; !exists {
			return padID(id), nil
		}
	}
	return "", errors.New("no available IDs")
}

func NewRecordKeeper(filePath string) *RecordKeeper {
	return &RecordKeeper{filePath: filePath}
}

// Add new record with given features
// Returns the ID
func (rk *RecordKeeper) Add(pid int) (string, error) {
	// Load store from disk
	rk.mu.Lock()
	defer rk.mu.Unlock()
	s, err := rk.loadStore()
	if err != nil {
		return "", err
	}

	// Get new ID
	id, err := nextAvailableID(s)
	if err != nil {
		return "", err
	}

	// Add record to store, save & return
	rec := Record{ID: id, PID: pid}
	s[id] = rec
	return id, rk.save(s)
}

// Remove deletes the given IDs; if ids is empty, removes ALL.
func (rk *RecordKeeper) Remove(ids []string) error {
	// Lock & load store
	rk.mu.Lock()
	defer rk.mu.Unlock()
	s, err := rk.loadStore()
	if err != nil {
		return err
	}

	// Delete records from store, save & return
	if len(ids) == 0 { // delete all records
		s = store{}
	}
	for _, id := range ids { // delete specific records
		delete(s, padID(id))
	}
	return rk.save(s)
}

// Get returns records for the given IDs in the same order.
// If ids is empty, returns all records sorted by ID ascending.
func (rk *RecordKeeper) Get(ids []string) ([]Record, error) {
	// Lock & load store
	rk.mu.Lock()
	defer rk.mu.Unlock()
	s, err := rk.loadStore()
	if err != nil {
		return nil, err
	}

	// Fetch IDs from store as list & return
	var out []Record
	if len(ids) == 0 { // fetch all records
		for _, v := range s {
			out = append(out, v)
		}
		return out, nil
	}
	for _, id := range ids { // fetch specific records
		if rec, ok := s[padID(id)]; ok {
			out = append(out, rec)
		}
	}
	return out, nil
}
