package commands

import (
	"fmt"
	"manager/utils"
)

func (a *App) Launch() error {
	id, err := a.Records.Add(-1)
	if err != nil {
		return fmt.Errorf("`launch` command failed: %v", err)
	}
	meta := utils.VMMetaData(id)
	fmt.Println(meta.IP())
	fmt.Println(meta.MacAddress())
	fmt.Println(meta.SocketPth())
	fmt.Println(meta.TapName())

	return nil
}
