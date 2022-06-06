package hive

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
)

type S3Client interface {
	PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

type SymlinkGenerator struct {
	s3client S3Client
}

func NewSymlinkGenerator(s3client S3Client) *SymlinkGenerator {
	return &SymlinkGenerator{
		s3client: s3client,
	}
}

func (sg *SymlinkGenerator) StoreSymlink(ctx context.Context, bucket, prefix string, hivePartitions, symlinkKeys []string) (*s3.PutObjectOutput, error) {

	key := fmt.Sprintf("%s/hive/%s/symlink.txt", prefix, strings.Join(hivePartitions, "/"))

	buf := bytes.NewBufferString(strings.Join(symlinkKeys, "\n"))

	putRes, err := sg.s3client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   buf,
	})
	if err != nil {
		return nil, err
	}

	log.Ctx(ctx).Info().Str("key", key).Str("id", aws.ToString(putRes.ChecksumSHA256)).Msg("put symlink")

	return putRes, nil
}
