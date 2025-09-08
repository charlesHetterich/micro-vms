package commands

import (
	"fmt"
	"manager/utils"
	"reflect"
	"strconv"
	"strings"
)

func header() string {
	rt := reflect.TypeOf(utils.Record{})
	headers := make([]string, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		headers[i] = rt.Field(i).Name
	}
	headers = append(headers, "STATUS")
	return strings.Join(headers, "\t")
}

func (a *App) List(ids []string) error {
	records, err := a.Records.Get(ids)
	if err != nil {
		return fmt.Errorf("`list` command failed: %w", err)
	}

	fmt.Println(header())
	for _, r := range records {
		pid := "-"
		if r.PID > 0 {
			pid = strconv.Itoa(r.PID)
		}
		fmt.Printf("%s\t%s\t%s\n", r.ID, pid, r.Status())
	}
	return nil
}
