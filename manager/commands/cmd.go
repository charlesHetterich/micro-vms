package commands

import (
	"fmt"
	"manager/utils"
	"os"
	"os/exec"
	"path/filepath"
)

// // shellQuote single-quotes one arg for a POSIX shell.
// func shellQuote(s string) string {
// 	// close ', insert '\'' , reopen '
// 	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
// }

// // buildRemoteCmd builds the command line to run under `sh -lc` on the guest.
// // If the user supplied ONE arg, we treat it as an already-quoted shell command
// // (so things like &&, |, redirections work when the user quotes it in their shell).
// // If multiple args were provided, we shell-quote each token to preserve spacing.
// func buildRemoteCmd(args []string) string {
// 	if len(args) == 1 {
// 		return args[0]
// 	}
// 	qs := make([]string, len(args))
// 	for i, a := range args {
// 		qs[i] = shellQuote(a)
// 	}
// 	return strings.Join(qs, " ")
// }

func (a *App) Cmd(id string, argv []string) error {
	if len(argv) == 0 {
		return fmt.Errorf("usage: manager cmd <id> <command ...>")
	}

	ip := utils.VMMetaData(id).IP()
	key := filepath.Join(os.Getenv("HOME"), ".ssh", "id_ed25519")

	sshArgs := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null", // avoids the “Permanently added …” line
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
