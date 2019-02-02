package s3cret

import (
	"errors"
	"strings"
)

type S3Object struct {
	Bucket string
	Key    string
	URL    string
}

// DeserializeS3URL ...
func DeserializeS3URL(url string) (*S3Object, error) {
	if len(url) < 1 {
		return nil, errors.New("url cannot have zero length")
	}

	const expectedPrefix = "s3://"
	urlPrefix := url[:len(expectedPrefix)]

	if !strings.EqualFold(urlPrefix, expectedPrefix) {
		return nil, errors.New("url must begin with 's3://'")
	}

	urlWithoutPrefix := url[len(expectedPrefix):]

	indexOfSlashAfterBucketName := strings.IndexAny(urlWithoutPrefix, "/")
	if indexOfSlashAfterBucketName == -1 {
		return nil, errors.New("url must include slash ('/') after bucket name")
	}

	bucket := urlWithoutPrefix[:indexOfSlashAfterBucketName]

	if len(urlWithoutPrefix)-1 <= len(bucket) {
		return nil, errors.New("url must include S3 key")
	}

	key := urlWithoutPrefix[(indexOfSlashAfterBucketName + 1):]

	return &S3Object{
		Bucket: bucket,
		Key:    key,
		URL:    url,
	}, nil
}
