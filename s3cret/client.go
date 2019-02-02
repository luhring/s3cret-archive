package s3cret

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/crypto/nacl/secretbox"
)

// Client ...
type Client struct {
	s3Client                 *s3.S3
	secretKey                *SecretKey
	encryptedDataChunkLength uint64
}

// NewClient ...
func NewClient(awsSession *session.Session) *Client {
	s3Client := s3.New(awsSession)
	secretKey := NewSecretKey()

	encryptedDataChunkLength := chunkSize + secretbox.Overhead

	return &Client{
		s3Client:                 s3Client,
		secretKey:                secretKey,
		encryptedDataChunkLength: encryptedDataChunkLength,
	}
}

func (c *Client) SendToS3(localPath, s3Bucket, s3Key string) error {
	fmt.Printf("overhead is %v bytes\n", secretbox.Overhead)

	secretKey := NewSecretKey()
	secretKey.Save()

	chunks := chunksFromFile(localPath)
	encryptedChunkBytes := c.encryptChunks(chunks)

	multipartUpload, err := newMultipartUpload(s3Bucket, s3Key, c.s3Client)
	if err != nil {
		return err
	}

	parts := c.createParts(encryptedChunkBytes, multipartUpload)

	completedParts := multipartUpload.uploadParts(parts)
	if completedParts == nil {
		multipartUpload.abort()
		return errors.New("had to abort multipart upload")
	}

	multipartUpload.complete(completedParts)
	return nil
}

func (c *Client) encryptChunks(chunks <-chan *chunk) <-chan []byte {
	encryptedChunks := make(chan []byte, 1)

	go func() {
		defer close(encryptedChunks)
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

			encryptedChunks <- encryptedChunk.toBytes()
		}
	}()
	return encryptedChunks
}

func (c *Client) createParts(
	byteChunks <-chan []byte,
	upload *multipartUpload,
) <-chan *s3.UploadPartInput {
	uploadableParts := make(chan *s3.UploadPartInput)

	go func() {
		defer close(uploadableParts)

		length := make([]byte, 8)
		binary.LittleEndian.PutUint64(length, uint64(c.encryptedDataChunkLength))

		var data [][]byte = nil
		var partNumber int64 = 1

		minimumCountOfChunksInPart := minimumUploadPartSize / chunkSize
		hasLastChunkBeenProcessed := false

		for hasLastChunkBeenProcessed == false {
			chk, isOpen := <-byteChunks
			if isOpen {
				data = append(data, chk)
			} else {
				hasLastChunkBeenProcessed = true
			}

			if uint64(len(data)) >= minimumCountOfChunksInPart || hasLastChunkBeenProcessed {
				bytesComponentsForReader := make([][]byte, len(data))
				copy(bytesComponentsForReader, data)
				bytesForReader := bytes.Join(bytesComponentsForReader, nil)
				r := bytes.NewReader(bytesForReader)

				part := &s3.UploadPartInput{
					Body:       r,
					Bucket:     aws.String(upload.bucket),
					Key:        aws.String(upload.key),
					PartNumber: aws.Int64(partNumber),
					UploadId:   aws.String(upload.uploadID),
				}

				uploadableParts <- part
				fmt.Printf("sent part %v to uploadableParts channel\n", *part.PartNumber)
				partNumber++
				data = nil
			}
		}
	}()
	return uploadableParts
}
