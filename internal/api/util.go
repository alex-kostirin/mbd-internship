package api

import (
	"encoding/binary"
	"github.com/golang/geo/s2"
)

func s2CellIdToKey(cellID s2.CellID) []byte {
	key := make([]byte, 8)
	binary.LittleEndian.PutUint64(key, uint64(cellID))
	return key
}
