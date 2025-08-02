package commands

import (
	"fmt"
	"strconv"
)

func (a *App) List(ids []string) error {
	records, err := a.Records.Get(ids)
	if err != nil {
		return fmt.Errorf("`list` command failed: %w", err)
	}
	fmt.Println("ID\tPID")
	for _, r := range records {
		pid := "-"
		if r.PID > 0 {
			pid = strconv.Itoa(r.PID)
		}
		fmt.Printf("%s\t%s\n", r.ID, pid)
	}
	return nil
}
