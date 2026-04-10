package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/restic/chunker"
)

func GetHashOfFile(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []string{}, err
	}
	defer file.Close()

	chunk := chunker.New(file, chunker.Pol(0x3DA3358B4DC173))

	buf := make([]byte, 8*1024*1024)

	var hashOfFile []string

	for {
		ch, err := chunk.Next(buf)

		if err == io.EOF {
			break
		}

		if err != nil {
			return []string{}, err
		}

		cache := sha256.Sum256(ch.Data)
		hashOfFile = append(hashOfFile, hex.EncodeToString(cache[:]))
	}

	return hashOfFile, nil
}
