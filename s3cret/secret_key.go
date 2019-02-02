package s3cret

import (
	"crypto/rand"
	"io"

	"github.com/satori/go.uuid"
)

type SecretKey struct {
	key [32]byte
	id  [16]byte
}

func (k *SecretKey) Save() {
	// TODO: implement save (just to disk, for now)
}

func generateSecretKey() ([32]byte, error) {
	var key [32]byte
	_, err := io.ReadFull(rand.Reader, key[:])
	return key, err
}

func NewSecretKey() *SecretKey {
	key, err := generateSecretKey()
	if err != nil {
		panic(err)
	}

	return &SecretKey{
		key,
		uuid.NewV4(),
	}
}
