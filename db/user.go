package db

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/leggettc18/hackernews-clone-api/model"
	"github.com/pkg/errors"
)

// GetUserByEmail returns the user with the specified email address from the database.
func (db *DB) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	if err := db.First(&user, model.User{Email: email}).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.Wrap(err, "unable to get user")
		}
	}
	return &user, nil
}

func (db *DB) GetUserById(id uint) (*model.User, error) {
	var user model.User
	if err := db.First(&user, model.User{ID: id}).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return nil, errors.Wrap(err, "unable to get user")
		}
	}
	return &user, nil
}

func (db *DB) GetUserFromToken(tokenString string) (*model.User, error) {
	// decode token with the secret it was encoded with
	tokenObj, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("verysecret"), nil
	})
	if err != nil {
		return nil, err
	}
	// get user ID from the map we encoded in the token
	userID, ok := tokenObj.Claims.(jwt.MapClaims)["ID"].(float64)
	if !ok {
		return nil, errors.New("GetUserIDFromToken error: type conversion in claims")
	}
	user, err := db.GetUserById(uint(userID))
	if err != nil {
		return nil, err
	}
	return user, nil
}

// CreateUser inserts a new user into the database.
func (db *DB) CreateUser(user *model.User) error {
	return db.Create(user).Error
}
