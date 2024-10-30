package models

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/minio/minio-go/v7"
)

// Create a struct that holds the client and bucket name
type BucketClient struct {
	Context context.Context
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
	return bc.Client.PutObject(bc.Context, bc.Bucket, objectName, reader, objectSize, opts)
}

func (bc *BucketClient) GetObject(objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return bc.Client.GetObject(bc.Context, bc.Bucket, objectName, opts)
}

func (bc *BucketClient) GetServiceObject(objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
	return bc.Client.GetObject(bc.Context, bc.ServiceBucket, objectName, opts)
}


func (bc *BucketClient) PutServiceObject(objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return bc.Client.PutObject(bc.Context, bc.ServiceBucket, objectName, reader, objectSize, opts)
}

func (bc *BucketClient) RemoveObject(objectName string, opts minio.RemoveObjectOptions) error {
	return bc.Client.RemoveObject(bc.Context, bc.Bucket, objectName, opts)
}

func (bc *BucketClient) RemoveServiceObject(objectName string, opts minio.RemoveObjectOptions) error {
	return bc.Client.RemoveObject(bc.Context, bc.ServiceBucket, objectName, opts)
}

func (bc *BucketClient) PresignedGetObject(objectName string, expires time.Duration, reqParams url.Values) (u *url.URL, err error) {
	return bc.Client.PresignedGetObject(bc.Context, bc.Bucket, objectName, expires, reqParams)
}