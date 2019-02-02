package s3cret

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"

	"golang.org/x/crypto/nacl/secretbox"
)

const bufferSizeForFileChunks = 100000

const chunkSize uint64 = 16384

type chunk struct {
	index int64
	data  []byte
}

func newChunk(data []byte, index int64) *chunk {
	return &chunk{
		index,
		data,
	}
}

func chunksFromFile(path string) <-chan *chunk {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("unable to open file: %v\n", err.Error())
	}

	chunks := make(chan *chunk, bufferSizeForFileChunks)

	go func() {
		defer close(chunks)

		var i int64 = 0
		for {
			data := make([]byte, chunkSize)
			_, err := f.Read(data)
			fmt.Printf("creating chunk w/ index %v\n", i)

			c := newChunk(data, i)
			chunks <- c

			if err != nil {
				if err != io.EOF {
					_, _ = fmt.Fprintf(os.Stderr, "error reading file: %v\n", err.Error())
				}
				return
			}

			i++
		}
	}()
	return chunks
}

func (c *chunk) encrypt(secretKey [32]byte) (*EncryptedChunk, error) {
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
