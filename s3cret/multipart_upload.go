package s3cret

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

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

	u, err := s3Client.CreateMultipartUpload(input)
	if err != nil {
		return nil, fmt.Errorf("error creating multipart upload: %v", err.Error())
	}

	return &multipartUpload{
		bucket:   bucket,
		key:      key,
		s3Client: s3Client,
		uploadID: *u.UploadId,
	}, nil
}

func (m *multipartUpload) abort() {
	abortInput := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(m.bucket),
		Key:      aws.String(m.key),
		UploadId: aws.String(m.uploadID),
	}

	fmt.Println("trying to abort multipart upload")
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

	fmt.Println("trying to complete multipart upload")
	_, err := m.s3Client.CompleteMultipartUpload(completeInput)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error completing multipart upload: %v\n", err.Error())
	} else {
		fmt.Printf("successfully completed multipart upload (id: %v)\n", m.uploadID)
	}
}

func (m *multipartUpload) uploadPart(part *s3.UploadPartInput) (*s3.CompletedPart, error) {
	result, err := m.s3Client.UploadPart(part)
	if err != nil {
		return nil, err
	}

	return &s3.CompletedPart{
		PartNumber: part.PartNumber,
		ETag:       result.ETag,
	}, nil
}

func (m *multipartUpload) uploadParts(
	uploadableParts <-chan *s3.UploadPartInput,
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

		fmt.Printf("successfully uploaded part number %v\n", *completedPart.PartNumber)
		completedParts = append(completedParts, completedPart)
	}

	return completedParts
}
