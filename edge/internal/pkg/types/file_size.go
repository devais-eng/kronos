package types

import (
	"github.com/dustin/go-humanize"
	"github.com/rotisserie/eris"
)

type FileSize uint64

func (f FileSize) String() string {
	return humanize.Bytes(uint64(f))
}

func (f FileSize) MarshalText() ([]byte, error) {
	return []byte(f.String()), nil
}

func FileSizeFromString(str string) (FileSize, error) {
	size, err := humanize.ParseBytes(str)
	if err != nil {
		return 0, eris.Wrap(err, "failed to parse file size")
	}
	return FileSize(size), nil
}
