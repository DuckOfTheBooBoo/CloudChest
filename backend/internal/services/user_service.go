package services

import (
	"context"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/gofrs/uuid/v5"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type UserService struct {
	DB *gorm.DB
	MinioClient *minio.Client
}

func (us *UserService) SetDB(db *gorm.DB) {
	us.DB = db
}

func (us *UserService) SetMinioClient(mc *minio.Client) {
	us.MinioClient = mc
}

func NewUserService(db *gorm.DB, mc *minio.Client) *UserService {
	return &UserService{
		DB: db,
		MinioClient: mc,
	}
}

func (us *UserService) CreateUser(userBody *models.UserBody) (*models.User, error) {
	ctx := context.Background()

	hashedPassword, err := utils.HashPassword(userBody.Password)

	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	// CREATE MINIO BUCKET
	bucketName, err := uuid.NewV4()
	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	serviceBucketName, err := uuid.NewV4()
	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	// Create user bucket
	err = us.MinioClient.MakeBucket(ctx, bucketName.String(), minio.MakeBucketOptions{
		Region: "us-east-1",
	})
	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	// Create service bucket
	err = us.MinioClient.MakeBucket(ctx, serviceBucketName.String(), minio.MakeBucketOptions{
		Region: "us-east-1",
	})
	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}
	
	rootFolder := models.Folder{
		Name: "/",
	}
	user := models.User{
		FirstName: userBody.FirstName,
		LastName: userBody.LastName,
		Email: userBody.Email,
		Password: hashedPassword,
		MinioBucket: bucketName.String(),
		MinioServiceBucket: serviceBucketName.String(),
		Folders: []*models.Folder{
			&rootFolder,
		},
	}

	err = us.DB.Create(&user).Error
	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	return &user, nil
}