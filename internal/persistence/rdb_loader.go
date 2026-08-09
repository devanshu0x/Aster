package persistence

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/devanshu0x/Aster/internal/store"
)

func readString(r io.Reader) (string, error) {
	var length uint32

	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return "", err
	}

	data := make([]byte, length)

	if _, err := io.ReadFull(r, data); err != nil {
		return "", err
	}

	return string(data), nil
}

func readObjectType(t uint8) (store.ObjType, error) {
	switch t {
	case stringType:
		return store.StringObject, nil
	case hashType:
		return store.HashObject, nil
	case listType:
		return store.ListObject, nil
	default:
		return "", fmt.Errorf("unknown RDB object type: %d", t)
	}
}

func readEntry(r io.Reader) (store.SnapshotEntry, error) {
	var entry store.SnapshotEntry

	// key
	key, err := readString(r)
	if err != nil {
		return entry, err
	}

	entry.Key = key

	// object type
	var objType uint8

	if err := binary.Read(r, binary.BigEndian, &objType); err != nil {
		return entry, err
	}

	entry.Type, err = readObjectType(objType)
	if err != nil {
		return entry, err
	}

	// expiration
	if err := binary.Read(r, binary.BigEndian, &entry.ExpiresAt); err != nil {
		return entry, err
	}

	// value
	switch entry.Type {
	case store.StringObject:
		value, err := readString(r)
		if err != nil {
			return entry, err
		}

		entry.Value = value

	default:
		return entry, fmt.Errorf(
			"deserialization not implemented for type: %s",
			entry.Type,
		)
	}

	return entry, nil
}


func LoadRDB(path string) (store.Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return store.Snapshot{}, err
	}
	defer f.Close()

	// Magic
	header := make([]byte, len(magic))

	if _, err := io.ReadFull(f, header); err != nil {
		return store.Snapshot{}, err
	}

	if string(header) != magic {
		return store.Snapshot{}, errors.New("invalid RDB magic")
	}

	// Version
	var fileVersion uint8

	if err := binary.Read(f, binary.BigEndian, &fileVersion); err != nil {
		return store.Snapshot{}, err
	}

	if fileVersion != version {
		return store.Snapshot{}, fmt.Errorf(
			"unsupported RDB version: %d",
			fileVersion,
		)
	}

	// Number of entries
	var entryCount uint64

	if err := binary.Read(f, binary.BigEndian, &entryCount); err != nil {
		return store.Snapshot{}, err
	}

	snapshot := store.Snapshot{
		Entries: make([]store.SnapshotEntry, 0, entryCount),
	}

	for i := uint64(0); i < entryCount; i++ {
		entry, err := readEntry(f)
		if err != nil {
			return store.Snapshot{}, fmt.Errorf(
				"failed reading entry %d: %w",
				i,
				err,
			)
		}

		snapshot.Entries = append(snapshot.Entries, entry)
	}

	return snapshot, nil
}
