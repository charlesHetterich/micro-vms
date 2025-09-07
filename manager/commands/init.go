package commands

import (
	"fmt"
	"manager/scripts"
	"os/exec"
)

func (a *App) Init() error {
	// PullRO (TODO! only if necessary)
	if err := scripts.Run("pullRO"); err != nil {
		return fmt.Errorf("failed to run connect script: %w", err)
	}

	// launch bridges and whatnot
	if err := exec.Command("ip", "link", "add", "lmbr0", "type", "bridge").Run(); err != nil {

	}

	return nil
}
