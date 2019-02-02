package s3cret

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const bufferSizeForEncryptedChunks = 100000

// Client ...
type Client struct {
	s3Client  *s3.S3
	secretKey *SecretKey
}

// NewClient ...
func NewClient(awsSession *session.Session) *Client {
	s3Client := s3.New(awsSession)
	secretKey := NewSecretKey()
	secretKey.Save()

	return &Client{
		s3Client:  s3Client,
		secretKey: secretKey,
	}
}

func (c *Client) SendToS3(localPath, s3Bucket, s3Key string) error {
	chunks := chunksFromFile(localPath)
	encryptedChunkBytes := c.encryptChunks(chunks)

	multipartUpload, err := newMultipartUpload(s3Bucket, s3Key, c.s3Client)
	if err != nil {
		return err
	}

	parts := multipartUpload.createParts(encryptedChunkBytes)

	completedParts := multipartUpload.uploadParts(parts)
	if completedParts == nil {
		multipartUpload.abort()
		return errors.New("had to abort multipart upload")
	}

	multipartUpload.complete(completedParts)
	return nil
}

func (c *Client) encryptChunks(chunks <-chan *chunk) <-chan []byte {
	encryptedBytes := make(chan []byte, bufferSizeForEncryptedChunks)

	go func() {
		defer close(encryptedBytes)
		for {
			chk, isOpen := <-chunks
			if !isOpen {
				fmt.Println("finished receiving chunks for encryption")
				return
			}

			encryptedChunk, err := chk.encrypt(c.secretKey.key)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "unable to encrypt chunk: %v\n", err.Error())
				return
			}

			encryptedBytes <- encryptedChunk.toBytes()
		}
	}()
	return encryptedBytes
}
