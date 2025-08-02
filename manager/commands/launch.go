package commands

import (
	"math/rand"
	"time"
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func (a *App) Launch() {
	const min, max = 1000, 65535
	pid := rng.Intn(max-min+1) + min
	a.Records.Add(pid)
}
