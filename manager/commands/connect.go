package commands

import (
	"fmt"
	"manager/utils"
	"os"
	"os/exec"
)

func (a *App) Connect(id string) error {
	cmd := exec.Command("ssh",
		"-tt",
		"-o", "StrictHostKeyChecking=no",
		"-i", os.Getenv("HOME")+"/.ssh/id_ed25519",
		"root@"+utils.VMMetaData(id).IP(),
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh: %w", err)
	}
	return nil
}
