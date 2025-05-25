package services

import (
	"context"
	"errors"
	"testing"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	// "github.com/gofrs/uuid/v5" // Not directly used in tests, but good to remember it's a dependency of the main code
)

// --- MockUserCreator ---
type MockUserCreator struct {
	mock.Mock
}

func (m *MockUserCreator) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// --- MockBucketCreator ---
type MockBucketCreator struct {
	mock.Mock
}

func (m *MockBucketCreator) MakeBucket(ctx context.Context, bucketName string, opts minio.MakeBucketOptions) error {
	args := m.Called(ctx, bucketName, opts)
	return args.Error(0)
}

func TestUserService_CreateUser(t *testing.T) {
	userBody := &models.UserBody{
		FirstName: "Test",
		LastName:  "User",
		Email:     "test@example.com",
		Password:  "password123",
	}

	t.Run("Successful User Creation", func(t *testing.T) {
		mockUserDB := new(MockUserCreator)
		mockMinio := new(MockBucketCreator)

		// Setup expectations
		// Expect MakeBucket to be called twice with any context, any string for bucketName, and any options. Both calls should succeed.
		mockMinio.On("MakeBucket", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("minio.MakeBucketOptions")).Return(nil).Twice()
		// Expect Create to be called once with any *models.User object. It should succeed.
		// We capture the argument to inspect it later.
		var capturedUser *models.User
		mockUserDB.On("Create", mock.MatchedBy(func(user *models.User) bool {
			capturedUser = user
			return true
		})).Return(nil).Once()

		userService := NewUserService(mockUserDB, mockMinio)
		createdUser, err := userService.CreateUser(userBody)

		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		if createdUser != nil { // Guard against nil dereference if assertion fails
			assert.Equal(t, userBody.FirstName, createdUser.FirstName)
			assert.Equal(t, userBody.LastName, createdUser.LastName)
			assert.Equal(t, userBody.Email, createdUser.Email)
			assert.NotEmpty(t, createdUser.Password) // Password should be hashed
			assert.NotEmpty(t, createdUser.MinioBucket, "MinioBucket should not be empty")
			assert.NotEmpty(t, createdUser.MinioServiceBucket, "MinioServiceBucket should not be empty")
			// Check if bucket names look like UUIDs (basic check for non-emptiness is already done)
			// A more robust check might involve trying to parse them as UUIDs, but not strictly necessary here.
			assert.Len(t, createdUser.MinioBucket, 36, "MinioBucket should be a UUID string")
			assert.Len(t, createdUser.MinioServiceBucket, 36, "MinioServiceBucket should be a UUID string")

			// Verify that the captured user passed to Create method has the generated bucket names
			assert.NotNil(t, capturedUser, "capturedUser should not be nil")
			if capturedUser != nil {
				assert.Equal(t, createdUser.MinioBucket, capturedUser.MinioBucket)
				assert.Equal(t, createdUser.MinioServiceBucket, capturedUser.MinioServiceBucket)
				assert.Equal(t, "/", capturedUser.Folders[0].Name) // Check for root folder
			}
		}

		mockMinio.AssertExpectations(t)
		mockUserDB.AssertExpectations(t)
	})

	t.Run("MinIO User Bucket Creation Fails", func(t *testing.T) {
		mockUserDB := new(MockUserCreator)
		mockMinio := new(MockBucketCreator)

		expectedErr := errors.New("minio make user bucket failed")
		// Setup mock to fail on the first call to MakeBucket
		mockMinio.On("MakeBucket", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("minio.MakeBucketOptions")).Return(expectedErr).Once()
		// No other calls to MakeBucket are expected
		// UserDB.Create should not be called

		userService := NewUserService(mockUserDB, mockMinio)
		createdUser, err := userService.CreateUser(userBody)

		assert.Error(t, err)
		assert.Nil(t, createdUser)

		var serverError *apperr.ServerError
		assert.ErrorAs(t, err, &serverError, "Error should be of type apperr.ServerError")
		if serverError != nil {
			assert.True(t, errors.Is(serverError.Err, expectedErr), "ServerError should wrap the original MinIO error")
			assert.Equal(t, "Internal server error ocurred", serverError.Message)
		}


		mockMinio.AssertExpectations(t)
		mockUserDB.AssertNotCalled(t, "Create", mock.Anything)
	})

	t.Run("MinIO Service Bucket Creation Fails", func(t *testing.T) {
		mockUserDB := new(MockUserCreator)
		mockMinio := new(MockBucketCreator)

		expectedErr := errors.New("minio make service bucket failed")
		// First call succeeds
		mockMinio.On("MakeBucket", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("minio.MakeBucketOptions")).Return(nil).Once()
		// Second call fails
		mockMinio.On("MakeBucket", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("minio.MakeBucketOptions")).Return(expectedErr).Once()
		// UserDB.Create should not be called

		userService := NewUserService(mockUserDB, mockMinio)
		createdUser, err := userService.CreateUser(userBody)

		assert.Error(t, err)
		assert.Nil(t, createdUser)

		var serverError *apperr.ServerError
		assert.ErrorAs(t, err, &serverError, "Error should be of type apperr.ServerError")
		if serverError != nil {
			assert.True(t, errors.Is(serverError.Err, expectedErr), "ServerError should wrap the original MinIO error")
			assert.Equal(t, "Internal server error ocurred", serverError.Message)
		}

		mockMinio.AssertExpectations(t)
		mockUserDB.AssertNotCalled(t, "Create", mock.Anything)
	})

	t.Run("Database User Creation Fails", func(t *testing.T) {
		mockUserDB := new(MockUserCreator)
		mockMinio := new(MockBucketCreator)

		expectedErr := errors.New("database create user failed")
		// Minio calls succeed
		mockMinio.On("MakeBucket", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("minio.MakeBucketOptions")).Return(nil).Twice()
		// DB call fails
		mockUserDB.On("Create", mock.AnythingOfType("*models.User")).Return(expectedErr).Once()

		userService := NewUserService(mockUserDB, mockMinio)
		createdUser, err := userService.CreateUser(userBody)

		assert.Error(t, err)
		assert.Nil(t, createdUser) // The user object is returned from DB, so it should be nil on error

		var serverError *apperr.ServerError
		assert.ErrorAs(t, err, &serverError, "Error should be of type apperr.ServerError")
		if serverError != nil {
			assert.True(t, errors.Is(serverError.Err, expectedErr), "ServerError should wrap the original DB error")
			assert.Equal(t, "Internal server error ocurred", serverError.Message)
		}

		mockMinio.AssertExpectations(t)
		mockUserDB.AssertExpectations(t)
	})
}
