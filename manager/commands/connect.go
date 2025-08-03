package commands

import (
	"fmt"
	"manager/scripts"
	"manager/utils"
)

func (a *App) Connect(id string) error {
	err := scripts.Run("connect", utils.VMMetaData(id).IP())
	if err != nil {
		return fmt.Errorf("failed to run connect script: %w", err)
	}
	return nil
}
