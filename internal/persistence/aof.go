package persistence

import (
	"os"

	"github.com/devanshu0x/Aster/internal/resp"
)

type AOF struct {
	file *os.File
}

func OpenAOF(path string) (*AOF, error) {
	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return nil, err
	}

	return &AOF{
		file: file,
	}, nil
}

func (a *AOF) Append(cmd *resp.RESPValue) error {
	data, err := resp.Encode(cmd)
	if err != nil {
		return err
	}

	_, err = a.file.Write(data)
	return err
}

func (a *AOF) Sync() error {
	return a.file.Sync()
}

func (a *AOF) Close() error {
	return a.file.Close()
}