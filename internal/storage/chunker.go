package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/restic/chunker"
)

func ProcessFileChunks(ctx context.Context, content io.Reader, processor *BlockProcessor, fileID int64) error {
	splitter := chunker.NewWithBoundaries(content, chunker.Pol(polynomial), minBlockSize, maxBlockSize)
	buf := make([]byte, bufSize)

	blockIndex := 0
	for {
		ch, err := splitter.Next(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if err := processor.ProcessBlock(ctx, ch.Data, fileID, blockIndex); err != nil {
			return fmt.Errorf("failed to process block: %d: %w", blockIndex, err)
		}
		blockIndex++
	}

	return nil
}
