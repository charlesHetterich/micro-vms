package commands

import "manager/utils"

type App struct {
	Records *utils.RecordKeeper
}

func NewApp(rk *utils.RecordKeeper) *App {
	return &App{Records: rk}
}
