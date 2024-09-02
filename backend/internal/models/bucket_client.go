package models

import (
	"context"
	"io"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/minio/minio-go/v7"
)

// Create a struct that holds the client and bucket name
type BucketClient struct {
	Client            *minio.Client
	Bucket        string
	ServiceBucket string
}

func NewBucketClient(minioClient *minio.Client, userClaim utils.UserClaims) *BucketClient {
	return &BucketClient{
		Client: minioClient,
		Bucket: userClaim.Bucket,
		ServiceBucket: userClaim.ServiceBucket,
	}
}

func (bc *BucketClient) PutObject(objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return bc.Client.PutObject(context.Background(), bc.Bucket, objectName, reader, objectSize, opts)
}

func (bc *BucketClient) GetObject(objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return bc.Client.GetObject(context.Background(), bc.Bucket, objectName, opts)
}