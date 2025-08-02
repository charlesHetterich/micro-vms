package utils

import (
	"fmt"
	c "manager/utils/constants"
	"path/filepath"
	"strconv"
)

type VMMetaData string

func (m VMMetaData) TapName() string {
	return "tap" + string(m)
}

func (m VMMetaData) SocketPth() string {
	return filepath.Join(c.TMP, string(m)+".sock")
}

func (m VMMetaData) IP() string {
	return "172.30.0." + string(m)
}

func (m VMMetaData) MacAddress() string {
	idNum, err := strconv.Atoi(string(m))
	if err != nil {
		panic("invalid vmId: " + string(m))
	}
	return fmt.Sprintf("AA:FC:00:00:00:%02X", idNum)
}
