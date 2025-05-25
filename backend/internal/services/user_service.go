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

// --- Interfaces ---

// UserCreator defines the interface for creating a user.
type UserCreator interface {
	Create(user *models.User) error
}

// BucketCreator defines the interface for creating a Minio bucket.
type BucketCreator interface {
	MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error
}

// --- Concrete Implementations ---

// GormUserCreator implements UserCreator using *gorm.DB.
type GormUserCreator struct {
	DB *gorm.DB
}

// NewGormUserCreator creates a new GormUserCreator.
func NewGormUserCreator(db *gorm.DB) *GormUserCreator {
	return &GormUserCreator{DB: db}
}

// Create creates a user record in the database.
func (guc *GormUserCreator) Create(user *models.User) error {
	return guc.DB.Create(user).Error
}

// MinioBucketCreator implements BucketCreator using *minio.Client.
type MinioBucketCreator struct {
	Client *minio.Client
}

// NewMinioBucketCreator creates a new MinioBucketCreator.
func NewMinioBucketCreator(client *minio.Client) *MinioBucketCreator {
	return &MinioBucketCreator{Client: client}
}

// MakeBucket creates a bucket in Minio.
func (mbc *MinioBucketCreator) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	return mbc.Client.MakeBucket(ctx, bucketName, opts)
}

// --- UserService ---

type UserService struct {
	UserDB      UserCreator    // Changed from DB *gorm.DB
	MinioClient BucketCreator  // Changed from MinioClient *minio.Client
}

func (us *UserService) SetDB(userDB UserCreator) { // Updated to accept UserCreator
	us.UserDB = userDB
}

func (us *UserService) SetMinioClient(mc BucketCreator) { // Updated to accept BucketCreator
	us.MinioClient = mc
}

func NewUserService(userDB UserCreator, mc BucketCreator) *UserService { // Updated parameters
	return &UserService{
		UserDB:      userDB,
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
	err = us.MinioClient.MakeBucket(ctx, bucketName.String(), minio.MakeBucketOptions{ // Use interface method
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
	err = us.MinioClient.MakeBucket(ctx, serviceBucketName.String(), minio.MakeBucketOptions{ // Use interface method
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

	err = us.UserDB.Create(&user) // Use interface method
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