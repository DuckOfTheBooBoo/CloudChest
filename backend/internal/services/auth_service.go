package services

import (
	"errors"
	"time"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

// UserFinder defines the interface for finding a user.
// It should behave similarly to gorm.DB's First method.
type UserFinder interface {
	First(out interface{}, where ...interface{}) error
}

// GormUserFinder is a concrete implementation of UserFinder for GORM.
type GormUserFinder struct {
	DB *gorm.DB
}

// First finds the first record that matches the given conditions, and populates 'out'.
func (guf *GormUserFinder) First(out interface{}, where ...interface{}) error {
	return guf.DB.First(out, where...).Error
}

type AuthService struct {
	UserFinder UserFinder
}

func NewAuthService(userFinder UserFinder) *AuthService {
	return &AuthService{
		UserFinder: userFinder,
	}
}

func (authS *AuthService) Login(email, password string) (string, error) {
	var user models.User
	if err := authS.UserFinder.First(&user, "email = ?", email); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "User not found",
					Err: err,
				},
			}
		}

		return "", &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	if !utils.CheckPassword(password, user.Password) {
		return "", &apperr.InvalidCredentialsError{
			BaseError: &apperr.BaseError{
				Message: "Invalid credentials",
			},
		}
	}

	userClaim := utils.UserClaims{
		ID: user.ID,
		Bucket: user.MinioBucket,
		ServiceBucket: user.MinioServiceBucket,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 60).Unix(),
		},
	}

	accessToken, err := utils.GenerateToken(userClaim)
	if err != nil {
		return "", &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	return accessToken, nil
}