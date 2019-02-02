package s3cret

import (
	"bytes"
	"encoding/binary"
)

type EncryptedChunk struct {
	index int64
	nonce [24]byte
	data  []byte
}

func (c *EncryptedChunk) toBytes() []byte {
	indexBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(indexBytes, uint64(c.index))

	components := [][]byte{
		indexBytes,
		c.nonce[:],
		c.data,
	}

	return bytes.Join(components, nil)
}
