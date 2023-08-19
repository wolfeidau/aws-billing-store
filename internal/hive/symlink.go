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

type StoreSymlinkResult struct {
	Bucket         string `json:"bucket"`
	Key            string `json:"key"`
	ChecksumSHA256 string `json:"checksum_sha_256"`
}

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

func (sg *SymlinkGenerator) StoreSymlink(ctx context.Context, bucket, prefix string, hivePartitions HivePartitions, symlinkKeys []string) (*StoreSymlinkResult, error) {

	key := fmt.Sprintf("%s/hive/%s/symlink.txt", prefix, hivePartitions.PathString())

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

	return &StoreSymlinkResult{
		Key:            key,
		Bucket:         bucket,
		ChecksumSHA256: aws.ToString(putRes.ChecksumSHA256),
	}, nil
}
