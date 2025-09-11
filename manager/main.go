package main

/*
TODO #1
- create `socket setup` go script which runs on vm
    - such that it is part of this project but is compilable separately from main (main doesn't compile it, it doesn't compile main. but they may share some of the other stuff)
- compile it to a binary
- make that binary available in cli tool (single overall binary)
- feed that binary into microvm base image on `init`
- cmd over socket from host (should speed up cmd execution)

TODO #2
- shortcut for cmd: `vm <id> <command ...>`
- normal is `vm cmd <id> <command ...>`

TODO #3
- snapshotting
- fast launch from snapshot (i.e. have a "zero" snapshot on `init`, and all launches are from that snapshot)

TODO #4
- analyze memory usage & see if any ways to reduce it
*/

import (
	"fmt"
	"manager/commands"
	"manager/utils"
	c "manager/utils/constants"
	"os"
	"os/exec"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./manager <launch|snap|launch_snap>")
		os.Exit(1)
	}

	app := commands.NewApp(utils.NewRecordKeeper(c.TMP + "/records.json"))

	// Shorthand: `vm <id> <command ...>` -> `vm cmd <id> <command ...>`
	if _, err := strconv.Atoi(os.Args[1]); err == nil {
		if len(os.Args) < 3 {
			fmt.Println("Usage: ./manager <id> <command ...>")
			os.Exit(1)
		}
		if err := app.Cmd(os.Args[1], os.Args[2:]); err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				os.Exit(ee.ExitCode())
			}
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

	cmd := os.Args[1]
	switch cmd {
	case "launch":
		if err := app.Launch(os.Args[2:]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "connect":
		if len(os.Args) != 3 {
			fmt.Println("Usage: ./manager connect <id>")
			os.Exit(1)
		}
		if err := app.Connect(os.Args[2]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "cmd":
		if len(os.Args) < 4 {
			fmt.Println("Usage: ./manager cmd <id> <command ...>")
			os.Exit(1)
		}
		if err := app.Cmd(os.Args[2], os.Args[3:]); err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				os.Exit(ee.ExitCode())
			}
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "cmd0":
		if len(os.Args) < 4 {
			fmt.Println("Usage: ./manager cmd0 <id> <command ...>")
			os.Exit(1)
		}
		if err := app.Cmd0(os.Args[2], os.Args[3:]); err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				os.Exit(ee.ExitCode())
			}
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "list":
		if err := app.List(os.Args[2:]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "delete":
		if err := app.Delete(os.Args[2:]); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
	case "init":
		if err := app.Init(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
