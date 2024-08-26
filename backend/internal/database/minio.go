package database

import (
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage interface {
	GetMinioClient() *minio.Client
}

type MinIO struct {
	client *minio.Client
}

func (client *MinIO) GetMinioClient() *minio.Client {
	return client.client
}

func ConnectToMinIO() (MinIOStorage, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		return nil, err
	}

	log.Println("Connected to MinIO")
	return &MinIO{client: minioClient}, nil
}
