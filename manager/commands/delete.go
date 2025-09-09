package commands

import (
	"errors"
	"fmt"
	"manager/utils"
	c "manager/utils/constants"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

func (a *App) Delete(ids []string) error {
	records, err := a.Records.Get(ids)
	if err != nil {
		return fmt.Errorf("delete records: %w", err)
	}

	var merr error
	for _, r := range records {
		if err := a.deleteResources(r); err != nil {
			merr = errors.Join(merr, fmt.Errorf("%s: %w", r.ID, err))
			continue
		}
		// remove record only after successful teardown of that VM
		if err := a.Records.Remove([]string{r.ID}); err != nil {
			merr = errors.Join(merr, fmt.Errorf("%s: remove record: %w", r.ID, err))
		}
	}
	return merr
}

// Clean up resources associated with a VM
//   - process
//   - tap device
//   - socket file
//   - overlay filesystem
func (a *App) deleteResources(r utils.Record) error {
	meta := utils.VMMetaData(r.ID)
	var merr error

	// Kill process
	if r.PID > 0 && processAlive(r.PID) {
		if err := killWithTimeout(r.PID, 3*time.Second); err != nil {
			merr = errors.Join(merr, fmt.Errorf("kill pid %d: %w", r.PID, err))
		}
	}

	// Delete socket & tap
	if err := os.Remove(meta.SocketPth()); err != nil && !os.IsNotExist(err) {
		merr = errors.Join(merr, fmt.Errorf("remove socket: %w", err))
	}
	if err := delTap(meta.TapName()); err != nil {
		merr = errors.Join(merr, fmt.Errorf("del tap: %w", err))
	}

	// 4) Unmount & cleanup overlay (if used) TODO! implement when we start using overlays
	if err := cleanOverlay(r.ID); err != nil {
		merr = errors.Join(merr, fmt.Errorf("overlay: %w", err))
	}
	return merr

}

func delTap(name string) error {
	// best-effort: down, detach from bridge, delete
	_ = exec.Command("ip", "link", "set", name, "down").Run()
	_ = exec.Command("ip", "link", "set", name, "nomaster").Run()
	if err := exec.Command("ip", "link", "delete", name).Run(); err != nil {
		// Only delete if it actually exists
		if exec.Command("ip", "link", "show", "dev", name).Run() == nil {
			return fmt.Errorf("ip link delete %s: %w", name, err)
		}
	}
	return nil
}

func processAlive(pid int) bool {
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

func killWithTimeout(pid int, grace time.Duration) error {
	_ = syscall.Kill(pid, syscall.SIGTERM)
	deadline := time.Now().Add(grace)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("pid %d did not exit", pid)
}

func cleanOverlay(vmId string) error {
	overlayDir := filepath.Join(c.TMP, vmId)
	if err := os.RemoveAll(overlayDir); err != nil {
		return fmt.Errorf("failed to remove overlay directory: %w", err)
	}
	return nil
}
