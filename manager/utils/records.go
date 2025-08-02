package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

const (
	minID = 5
	maxID = 250
)

type Record struct {
	ID  string `json:"id"`            // 5..250
	PID int    `json:"pid,omitempty"` // optional
}

type RecordKeeper struct {
	filePath string
	mu       sync.Mutex
}

func NewRecordKeeper(filePath string) *RecordKeeper {
	return &RecordKeeper{filePath: filePath}
}

func padID(id string) string {
	if len(id) < 3 {
		return fmt.Sprintf("%03s", id)
	}
	return id
}

func validateID(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("invalid id %q: %w", id, err)
	}
	if idInt < minID || idInt > maxID {
		return fmt.Errorf("id %d out of range [%d,%d]", idInt, minID, maxID)
	}
	return nil
}

func (rk *RecordKeeper) ensureFile() error {
	dir := filepath.Dir(rk.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// If file doesn’t exist, create an empty JSON object.
	if _, err := os.Stat(rk.filePath); errors.Is(err, os.ErrNotExist) {
		return os.WriteFile(rk.filePath, []byte("{}\n"), 0o644)
	}
	return nil
}

type store = map[string]Record

func (rk *RecordKeeper) load() (store, error) {
	if err := rk.ensureFile(); err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(rk.filePath)
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return store{}, nil
	}
	var s store
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("parse %s: %w", rk.filePath, err)
	}
	if s == nil {
		s = store{}
	}
	return s, nil
}

func (rk *RecordKeeper) save(s store) error {
	tmp := rk.filePath + ".tmp"
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	// Atomic on POSIX: replaces the old file in one step.
	return os.Rename(tmp, rk.filePath)
}

// nextAvailableID finds the lowest unused ID in [minID, maxID].
func nextAvailableID(s store) (string, error) {
	for idRaw := minID; idRaw <= maxID; idRaw++ {
		id := strconv.Itoa(idRaw)
		if _, exists := s[padID(id)]; !exists {
			return padID(id), nil
		}
	}
	return "", errors.New("no available IDs")
}

// Add creates a new record with the lowest available ID and the provided features.
// Currently the only feature is PID (optional). Returns the created Record.
func (rk *RecordKeeper) Add(pid int) (Record, error) {
	rk.mu.Lock()
	defer rk.mu.Unlock()

	s, err := rk.load()
	if err != nil {
		return Record{}, err
	}

	id, err := nextAvailableID(s)
	if err != nil {
		return Record{}, err
	}
	if err := validateID(id); err != nil {
		return Record{}, err
	}

	rec := Record{ID: id, PID: pid}
	s[id] = rec

	if err := rk.save(s); err != nil {
		return Record{}, err
	}
	return rec, nil
}

// Remove deletes the given IDs; if ids is empty, removes ALL.
func (rk *RecordKeeper) Remove(ids []string) (err error) {
	rk.mu.Lock()
	defer rk.mu.Unlock()

	s, err := rk.load()
	if err != nil {
		return err
	}

	if len(ids) == 0 {
		s = store{}
		return rk.save(s)
	}

	for _, id := range ids {
		if err := validateID(id); err != nil {
			return err
		}
		delete(s, id)
	}
	return rk.save(s)
}

// Get returns records for the given IDs in the same order.
// If ids is empty, returns all records sorted by ID ascending.
func (rk *RecordKeeper) Get(ids []string) ([]Record, error) {
	rk.mu.Lock()
	defer rk.mu.Unlock()
	s, err := rk.load()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		// all, sorted by numeric ID
		keys := make([]string, 0, len(s))
		for k := range s {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out := make([]Record, 0, len(keys))
		for _, k := range keys {
			out = append(out, s[k])
		}
		return out, nil
	}

	out := make([]Record, 0, len(ids))
	for _, id := range ids {
		id = padID(id) // ensure ID is padded
		if err := validateID(id); err != nil {
			return nil, err
		}
		if rec, ok := s[id]; ok {
			out = append(out, rec)
		}
	}
	return out, nil
}
