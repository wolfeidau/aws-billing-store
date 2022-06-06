package hive

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/require"
)

type MockS3Client struct {
	putObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

func (m *MockS3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m.putObjectFunc(ctx, params, optFns...)
}

func TestSymlinkGenerator_StoreSymlink(t *testing.T) {

	assert := require.New(t)

	type args struct {
		ctx            context.Context
		bucket         string
		prefix         string
		hivePartitions []string
		symlinkKeys    []string
	}
	tests := []struct {
		name          string
		args          args
		wantBucket    string
		wantKey       string
		wantBody      string
		wantPutObject *s3.PutObjectOutput
		wantErr       bool
	}{
		{
			name: "should create symlink with valid input",
			args: args{
				ctx:            context.TODO(),
				bucket:         "test-bucket",
				prefix:         "test-prefix",
				hivePartitions: []string{"year=2022", "month=1"},
				symlinkKeys: []string{
					"parquet/test-managment-cur/20220401-20220501/20220503T120125Z/test-managment-cur-00001.snappy.parquet",
					"parquet/test-managment-cur/20220401-20220501/20220503T120125Z/test-managment-cur-00002.snappy.parquet",
				},
			},
			wantBucket:    "test-bucket",
			wantKey:       "test-prefix/hive/year=2022/month=1/symlink.txt",
			wantBody:      "parquet/test-managment-cur/20220401-20220501/20220503T120125Z/test-managment-cur-00001.snappy.parquet\nparquet/test-managment-cur/20220401-20220501/20220503T120125Z/test-managment-cur-00002.snappy.parquet",
			wantPutObject: &s3.PutObjectOutput{},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := &SymlinkGenerator{
				s3client: &MockS3Client{
					putObjectFunc: func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
						assert.Equal(tt.wantBucket, aws.ToString(params.Bucket))
						assert.Equal(tt.wantKey, aws.ToString(params.Key))

						data, err := io.ReadAll(params.Body)
						assert.NoError(err)

						fmt.Println(string(data))

						assert.Equal(tt.wantBody, string(data))

						return &s3.PutObjectOutput{}, nil
					},
				},
			}
			got, err := sg.StoreSymlink(tt.args.ctx, tt.args.bucket, tt.args.prefix, tt.args.hivePartitions, tt.args.symlinkKeys)
			if tt.wantErr {
				assert.Error(err)
			}

			assert.Equal(tt.wantPutObject, got)
		})
	}
}
