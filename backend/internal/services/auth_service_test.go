package services

import (
	"errors"
	"testing"
	// "time" // Removed as it's unused

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// MockUserFinder is a mock implementation of the UserFinder interface.
type MockUserFinder struct {
	UserToReturn  *models.User
	ErrorToReturn error
}

// First simulates the First method of UserFinder.
func (m *MockUserFinder) First(out interface{}, where ...interface{}) error {
	if m.ErrorToReturn != nil {
		return m.ErrorToReturn
	}
	if m.UserToReturn != nil {
		if user, ok := out.(*models.User); ok {
			*user = *m.UserToReturn
		} else {
			return errors.New("mockUserFinder: 'out' is not *models.User")
		}
		return nil
	}
	return gorm.ErrRecordNotFound
}

func TestAuthService_Login(t *testing.T) {
	hashedPassword, _ := utils.HashPassword("correctpassword")
	// This password is intentionally different from "correctpassword" to test incorrect password scenario
	incorrectHashedPasswordForTestUser, _ := utils.HashPassword("otherpassword")

	tests := []struct {
		name           string
		email          string
		password       string
		mockUserFinder UserFinder
		expectedToken  bool
		expectedError  error // Store the expected error type
		expectedErrMsg string // Store the expected error message if specific
	}{
		{
			name:     "Successful Login",
			email:    "test@example.com",
			password: "correctpassword",
			mockUserFinder: &MockUserFinder{
				UserToReturn: &models.User{
					Model:              gorm.Model{ID: 1}, // Correctly set ID via gorm.Model
					Email:              "test@example.com",
					Password:           hashedPassword,
					MinioBucket:        "test-bucket",
					MinioServiceBucket: "test-service-bucket",
				},
			},
			expectedToken: true,
		},
		{
			name:     "User Not Found",
			email:    "nosuchuser@example.com",
			password: "password",
			mockUserFinder: &MockUserFinder{
				ErrorToReturn: gorm.ErrRecordNotFound,
			},
			expectedToken: false,
			expectedError: &apperr.NotFoundError{},
			expectedErrMsg: "User not found",
		},
		{
			name:     "Incorrect Password",
			email:    "test@example.com",
			password: "wrongpassword", // User provides this password
			mockUserFinder: &MockUserFinder{
				UserToReturn: &models.User{
					Model:              gorm.Model{ID: 2},
					Email:              "test@example.com",
					Password:           incorrectHashedPasswordForTestUser, // User in DB has "otherpassword"
					MinioBucket:        "test-bucket",
					MinioServiceBucket: "test-service-bucket",
				},
			},
			expectedToken: false,
			expectedError: &apperr.InvalidCredentialsError{},
			expectedErrMsg: "Invalid credentials",
		},
		{
			name:     "Database Error on Find",
			email:    "anyuser@example.com",
			password: "anypassword",
			mockUserFinder: &MockUserFinder{
				ErrorToReturn: errors.New("database exploded"),
			},
			expectedToken: false,
			expectedError: &apperr.ServerError{},
			expectedErrMsg: "Internal server error ocurred",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			authService := NewAuthService(tc.mockUserFinder)
			token, err := authService.Login(tc.email, tc.password)

			if tc.expectedToken {
				assert.NotEmpty(t, token, "Expected a token but got none")
			} else {
				assert.Empty(t, token, "Expected no token but got one")
			}

			if tc.expectedError != nil {
				assert.Error(t, err, "Expected an error but got none")
				assert.IsType(t, tc.expectedError, err, "Error type mismatch")

				// Check message if specified
				if tc.expectedErrMsg != "" {
					// Access BaseError's Message field after type assertion
					switch e := err.(type) {
					case *apperr.NotFoundError:
						assert.Equal(t, tc.expectedErrMsg, e.BaseError.Message, "Error message mismatch")
					case *apperr.InvalidCredentialsError:
						assert.Equal(t, tc.expectedErrMsg, e.BaseError.Message, "Error message mismatch")
					case *apperr.ServerError:
						assert.Equal(t, tc.expectedErrMsg, e.BaseError.Message, "Error message mismatch")
					default:
						// If it's a generic error that doesn't fit apperr types but is still expected
						assert.Equal(t, tc.expectedErrMsg, err.Error(), "Generic error message mismatch")
					}
				}
			} else {
				assert.NoError(t, err, "Expected no error but got one")
			}
		})
	}
}
