package persistence

import (
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/devanshu0x/Aster/internal/store"
)

func writeString(w io.Writer, s string) error {
	if err := binary.Write(w, binary.BigEndian, uint32(len(s))); err != nil {
		return err
	}

	_, err := io.WriteString(w, s)
	return err
}

func writeEntry(w io.Writer, entry store.SnapshotEntry) error {
	// key
	if err := writeString(w, entry.Key); err != nil {
		return err
	}

	// expiration timestamp
	if err := binary.Write(w, binary.BigEndian, entry.ExpiresAt); err != nil {
		return err
	}

	
		value, ok := entry.Value.(string)
		if !ok {
			return errors.New("string object contains non-string value")
		}

		return writeString(w, value)

	
}

func SaveRDB(snapshot store.Snapshot, path string) error {
	tmpPath := path + ".tmp"

	f, err := os.OpenFile(
		tmpPath,
		os.O_CREATE|os.O_WRONLY|os.O_TRUNC,
		0644,
	)
	if err != nil {
		return err
	}

	success := false

	defer func() {
		f.Close()

		if !success {
			os.Remove(tmpPath)
		}
	}()

	// Magic
	if _, err := io.WriteString(f, magic); err != nil {
		return err
	}

	// Version
	if err := binary.Write(f, binary.BigEndian, version); err != nil {
		return err
	}

	// Number of entries
	entryCount := uint64(len(snapshot.Entries))

	if err := binary.Write(f, binary.BigEndian, entryCount); err != nil {
		return err
	}

	// Entries
	for _, entry := range snapshot.Entries {
		if err := writeEntry(f, entry); err != nil {
			return err
		}
	}

	// Make sure the RDB data reaches the filesystem.
	if err := f.Sync(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	// Atomically replace the previous RDB.
	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	success = true

	return nil
}