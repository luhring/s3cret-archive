package s3cret

import (
	"bytes"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const minimumUploadPartSize uint64 = 5.243e+6

type part struct {
	partNumber int64
	body       []byte
	bucket     string
	key        string
	uploadID   string
}

func (p *part) toS3UploadPartInput() *s3.UploadPartInput {
	return &s3.UploadPartInput{
		Body:       bytes.NewReader(p.body),
		Bucket:     aws.String(p.bucket),
		Key:        aws.String(p.key),
		PartNumber: aws.Int64(p.partNumber),
		UploadId:   aws.String(p.uploadID),
	}
}

func (p *part) isLargeEnoughForUpload() bool {
	return uint64(len(p.body)) >= minimumUploadPartSize
}

func (p *part) mergeWith(secondPart *part) *part {
	var mergedBodyComponents [][]byte = nil
	var resultingPartNumber int64

	if p.partNumber < secondPart.partNumber {
		mergedBodyComponents = [][]byte{
			p.body,
			secondPart.body,
		}
		resultingPartNumber = p.partNumber
	} else {
		mergedBodyComponents = [][]byte{
			secondPart.body,
			p.body,
		}
		resultingPartNumber = secondPart.partNumber
	}

	return &part{
		partNumber: resultingPartNumber,
		body:       bytes.Join(mergedBodyComponents, nil),
		bucket:     p.bucket,
		key:        p.key,
		uploadID:   p.uploadID,
	}
}

func doesPartHaveEnoughChunks(chunkCount uint64) bool {
	minimumCountOfChunksInPart := minimumUploadPartSize / chunkSize
	return chunkCount >= minimumCountOfChunksInPart
}
