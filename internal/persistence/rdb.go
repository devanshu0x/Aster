package persistence

import (
	"fmt"

	"github.com/devanshu0x/Aster/internal/store"
)

const (
	magic   = "ASTERDB"
	version = uint8(1)

	stringType uint8 = 1
	hashType   uint8 = 2
	listType   uint8 = 3
)


func objectType(t store.ObjType) (uint8, error) {
	switch t {
	case store.StringObject:
		return stringType, nil
	case store.HashObject:
		return hashType, nil
	case store.ListObject:
		return listType, nil
	default:
		return 0, fmt.Errorf("unsupported object type: %s", t)
	}
}
