package commands

import (
	"fmt"
	"manager/utils"
	c "manager/utils/constants"
	"os"
	"os/exec"
	"path/filepath"
)

func (a *App) Cmd0(id string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("usage: manager cmd <id> <command ...>")
	}

	ip := utils.VMMetaData(id).IP()
	key := filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519")
	kh := filepath.Join(c.TMP, "known_hosts")

	sshArgs := []string{
		"-o", "StrictHostKeyChecking=accept-new", // adds without prompt; fine for throwaway VMs
		"-o", "UserKnownHostsFile=" + kh, // avoids the “Permanently added …” line
		"-o", "GlobalKnownHostsFile=/dev/null", // ditto
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=3",
		"-i", key,
		"root@" + ip,
	}
	// Pass the command *as tokens*; ssh will quote/escape for the remote shell.
	sshArgs = append(sshArgs, argv...)

	cmd := exec.Command("ssh", sshArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run() // main.go will map *exec.ExitError to exit code
}
