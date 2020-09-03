package history

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// S3Client is an abstraction layer for AWS S3 API.
// It is intended to be replaced with a mock for testing.
type S3Client interface {
	// PutObjectWithContext puts a file to S3.
	PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error)
	// GetObjectWithContext gets a file from S3.
	GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error)
}

// s3Client is a real implemention of S3Client.
type s3Client struct {
	s3api s3iface.S3API
}

// PutObjectWithContext puts a file to S3.
func (c *s3Client) PutObjectWithContext(ctx aws.Context, input *s3.PutObjectInput, opts ...request.Option) (*s3.PutObjectOutput, error) {
	return c.s3api.PutObjectWithContext(ctx, input, opts...)
}

// GetObjectWithContext gets a file from S3.
func (c *s3Client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	return c.s3api.GetObjectWithContext(ctx, input, opts...)
}

// S3Storage is an implementation of Storage for AWS S3.
type S3Storage struct {
	// Client is an instance of S3Client interface to call API.
	// It is intended to be replaced with a mock for testing.
	client S3Client
	// Bucket is a name of s3 bucket.
	bucket string
	// Key is a path to the migration history file.
	key string
}

var _ Storage = (*S3Storage)(nil)

// NewS3Storage returns a new instance of S3Storage.
func NewS3Storage(client S3Client, bucket string, key string) (*S3Storage, error) {
	s := &S3Storage{
		client: client,
		bucket: bucket,
		key:    key,
	}

	return s, nil
}

// NewS3Client returns a new instance of S3Client.
func NewS3Client() (S3Client, error) {
	sess, err := session.NewSession()
	if err != nil {
		return nil, err
	}
	client := &s3Client{
		s3api: s3.New(sess),
	}
	return client, nil
}

// Write writes migration history data to storage.
func (s *S3Storage) Write(ctx context.Context, b []byte) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
		Body:   bytes.NewReader(b),
	}

	_, err := s.client.PutObjectWithContext(ctx, input)

	return err
}

// Read reads migration history data from storage.
func (s *S3Storage) Read(ctx context.Context) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(s.key),
	}

	output, err := s.client.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, err
	}

	defer output.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err = buf.ReadFrom(output.Body)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
