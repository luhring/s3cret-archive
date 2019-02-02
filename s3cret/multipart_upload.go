package s3cret

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const bufferSizeForUploadParts = 100

type multipartUpload struct {
	bucket   string
	key      string
	s3Client *s3.S3
	uploadID string
}

func newMultipartUpload(bucket string, key string, s3Client *s3.S3) (*multipartUpload, error) {
	input := &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	m, err := s3Client.CreateMultipartUpload(input)
	if err != nil {
		return nil, fmt.Errorf("error creating multipart upload: %v", err.Error())
	}

	return &multipartUpload{
		bucket:   bucket,
		key:      key,
		s3Client: s3Client,
		uploadID: *m.UploadId,
	}, nil
}

func (m *multipartUpload) abort() {
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(m.bucket),
		Key:      aws.String(m.key),
		UploadId: aws.String(m.uploadID),
	}

	_, err := m.s3Client.AbortMultipartUpload(abortInput)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error aborting multipart upload: %v\n", err.Error())
	} else {
		fmt.Print("successfully aborted multipart upload\n")
	}

	return
}

func (m *multipartUpload) complete(completedParts []*s3.CompletedPart) {
	completeInput := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(m.bucket),
		Key:      aws.String(m.key),
		UploadId: aws.String(m.uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}

	fmt.Printf("manifest of mulipart upload parts:\n%v\n", completedParts)

	_, err := m.s3Client.CompleteMultipartUpload(completeInput)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error completing multipart upload: %v\n", err.Error())
	} else {
		fmt.Printf("successfully completed multipart upload (id: %v)\n", m.uploadID)
	}
}

func (m *multipartUpload) createPart(byteChunks <-chan []byte, partNumber int64) *part {
	var data [][]byte = nil

	for chunkCount := uint64(0); !doesPartHaveEnoughChunks(chunkCount); chunkCount++ {
		chk, isOpen := <-byteChunks
		if !isOpen {
			if chunkCount == 0 {
				return nil // can't create a part
			}

			break
		}

		data = append(data, chk)
		fmt.Printf("added chunk index %v to part number %v\n", chunkCount, partNumber)
	}

	bytesComponentsForReader := make([][]byte, len(data))
	copy(bytesComponentsForReader, data)

	return &part{
		body:       bytes.Join(bytesComponentsForReader, nil),
		bucket:     m.bucket,
		key:        m.key,
		partNumber: partNumber,
		uploadID:   m.uploadID,
	}
}

func (m *multipartUpload) createParts(
	byteChunks <-chan []byte,
) <-chan *part {
	parts := make(chan *part, bufferSizeForUploadParts)

	go func() {
		defer close(parts)

		var previousPart *part = nil

		partNumber := int64(1)
		for {
			p := m.createPart(byteChunks, partNumber)
			if p == nil {
				break
			}

			// if part is too small, merge with previous part
			if !p.isLargeEnoughForUpload() {
				if previousPart == nil {
					panic("unable to produce parts large enough for upload")
				}

				p := p.mergeWith(previousPart)
				parts <- p
				fmt.Printf("sent part %v (with part %v merged into it) to parts channel\n", p.partNumber, partNumber)

				break
			}

			if previousPart != nil {
				parts <- previousPart
				fmt.Printf("sent part %v to parts channel\n", previousPart.partNumber)
			}

			partNumber++
			previousPart = p
		}
	}()
	return parts
}

func (m *multipartUpload) uploadPart(part *part) (*s3.CompletedPart, error) {
	partForS3 := part.toS3UploadPartInput()

	fmt.Printf("uploading part %v...\n", part.partNumber)
	result, err := m.s3Client.UploadPart(partForS3)
	if err != nil {
		return nil, err
	}

	fmt.Printf("uploaded part %v\n", part.partNumber)

	return &s3.CompletedPart{
		PartNumber: aws.Int64(part.partNumber),
		ETag:       result.ETag,
	}, nil
}

func (m *multipartUpload) uploadParts(
	uploadableParts <-chan *part,
) []*s3.CompletedPart {
	var completedParts []*s3.CompletedPart

	for {
		part, isOpen := <-uploadableParts
		if !isOpen {
			fmt.Println("finished receiving parts for upload")
			break
		}

		completedPart, err := m.uploadPart(part)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error uploading chunk: %v\n", err.Error())
			return nil
		}

		fmt.Printf("successfully uploaded part %v\n", *completedPart.PartNumber)
		completedParts = append(completedParts, completedPart)
	}

	return completedParts
}
