package models

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
)

// Create a struct that holds the client and bucket name
type BucketClient struct {
    Client     *minio.Client
    BucketName string
}

func (bc *BucketClient) PutObject(objectName string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
    return bc.Client.PutObject(context.Background(), bc.BucketName, objectName, reader, objectSize, opts)
}

func (bc *BucketClient) GetObject(objectName string, opts minio.GetObjectOptions) (*minio.Object, error) {
    return bc.Client.GetObject(context.Background(), bc.BucketName, objectName, opts)
}
