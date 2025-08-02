package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type VMIds struct {
	filePath string
	mu       sync.Mutex
}

func NewVMIds(filePath string) *VMIds {
	return &VMIds{filePath: filePath}
}

// AddID now takes an id and a pid
func (v *VMIds) AddID(id string, pid int) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	records := v.getRecords()
	for _, rec := range records {
		if rec.id == id {
			return nil // already exists
		}
	}
	f, err := os.OpenFile(v.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(id + "," + strconv.Itoa(pid) + "\n")
	return err
}

func (v *VMIds) RmRecords(ids []string) ([]int, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	// If no IDs given, truncate the file (delete all)
	if len(ids) == 0 {
		var pids []int
		for _, rec := range v.getRecords() {
			pids = append(pids, rec.pid)
		}
		if err := os.Truncate(v.filePath, 0); err != nil {
			return nil, err
		}
		return pids, nil
	}

	// Build a set for efficient lookup
	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
	}

	var deletedPIDs []int
	records := v.getRecords()
	f, err := os.OpenFile(v.filePath, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	defer f.Close()
	for _, rec := range records {
		if _, remove := idSet[rec.id]; !remove {
			_, err := f.WriteString(rec.id + "," + strconv.Itoa(rec.pid) + "\n")
			if err != nil {
				return nil, err
			}
		}
	}
	return deletedPIDs, nil
}

func (v *VMIds) GetRecords() []vmRecord {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.getRecords()
}

// --- helper struct and func ---

type vmRecord struct {
	id  string
	pid int
}

// Reads all id,pid pairs from file
func (v *VMIds) getRecords() []vmRecord {
	f, err := os.OpenFile(v.filePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil
	}
	defer f.Close()
	var records []vmRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 2)
		if len(parts) != 2 {
			continue // skip malformed
		}
		id := parts[0]
		pid, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		records = append(records, vmRecord{id, pid})
	}
	return records
}

func (v *VMIds) GetAvailableIp() (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	used := make(map[int]struct{})
	for _, rec := range v.getRecords() {
		if n, err := strconv.Atoi(rec.id); err == nil {
			used[n] = struct{}{}
		}
	}
	for i := 2; i <= 254; i++ {
		if _, taken := used[i]; !taken {
			return strconv.Itoa(i), nil
		}
	}
	return "", fmt.Errorf("no available IPs in range 2-254")
}

func (v *VMIds) GetSocketPth(id string) string {
	return filepath.Join(TMP_DIR, id+".sock")
}

func (v *VMIds) GetTapName(id string) string {
	return "tap" + id
}

func (v *VMIds) GetIp(id string) string {
	return "172.30.0." + id
}

func (v *VMIds) GetMacAddress(id string) string {
	idNum, err := strconv.Atoi(id)
	if err != nil {
		panic("invalid vmId: " + id)
	}
	return fmt.Sprintf("AA:FC:00:00:00:%02X", idNum)
}
