package s3cret

import (
	"crypto/rand"
	"fmt"
	"io"

	"golang.org/x/crypto/nacl/secretbox"
)

type Chunk struct {
	index int64
	data  []byte
}

func (c *Chunk) encrypt(secretKey [32]byte) (*EncryptedChunk, error) {
	var nonce [24]byte
	if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
		return nil, err
	}

	fmt.Printf("encrypting chunk w/ index %v\n", c.index)

	sealedData := secretbox.Seal(nil, c.data, &nonce, &secretKey)

	encryptedChunk := &EncryptedChunk{
		index: c.index,
		data:  sealedData,
		nonce: nonce,
	}

	return encryptedChunk, nil
}
