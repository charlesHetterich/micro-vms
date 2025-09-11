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
	idNum, err := strconv.Atoi(string(m))
	if err != nil {
		panic("invalid vmId: " + string(m))
	}
	return "172.30.0." + strconv.Itoa(idNum)
}

func (m VMMetaData) MacAddress() string {
	idNum, err := strconv.Atoi(string(m))
	if err != nil {
		panic("invalid vmId: " + string(m))
	}
	return fmt.Sprintf("AA:FC:00:00:00:%02X", idNum)
}

// CID returns a unique vsock CID per VM (>=3, since 0/1/2 are reserved).
func (m VMMetaData) CID() uint32 {
	idNum, err := strconv.Atoi(string(m))
	if err != nil {
		panic("invalid vmId: " + string(m))
	}
	return 1000 + uint32(idNum) // pick any range >=3; 1000+ID is easy to eyeball
}

// (Optional) If your SDK version needs a host UDS path for vsock backend:
func (m VMMetaData) VsockUDS() string {
	return filepath.Join(c.TMP, "vsock-"+string(m)+".sock")
}
