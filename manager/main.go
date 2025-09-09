package main

import (
	"fmt"
	"manager/commands"
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
