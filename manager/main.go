package main

import (
	"fmt"
	"manager/commands"
	"manager/scripts"
	"manager/utils"
	c "manager/utils/constants"
	"os"
	// fcSdk "github.com/firecracker-microvm/firecracker-go-sdk"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./manager <launch|snap|launch_snap>")
		os.Exit(1)
	}

	app := commands.NewApp(utils.NewRecordKeeper(c.TMP + "/records.json"))

	cmd := os.Args[1]
	switch cmd {
	case "launch":
		app.Launch()
	case "connect":
		if len(os.Args) != 3 {
			fmt.Println("Usage: ./manager connect <id>")
			os.Exit(1)
		}
		app.Connect(os.Args[2])
	case "list":
		app.List(os.Args[2:])
	case "delete":
		app.Delete(os.Args[2:])
	case "init":
		scripts.Run("pullRO")
	default:
		fmt.Println("Unknown command:", cmd)
		os.Exit(1)
	}
}
