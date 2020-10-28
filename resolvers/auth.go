package resolvers

import (
	"context"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type AuthResolver struct {
	AuthPayload AuthPayload
}

type AuthPayload struct {
	Token *string
	User  *User
}

func GetUserFromToken(tokenString string) (*User, error) {
	// decode token with the secret it was encoded with
	tokenObj, err := jwt.ParseWithClaims(tokenString, jwt.MapClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("verysecret"), nil
	})
	if err != nil {
		return nil, err
	}
	// get user ID from the map we encoded in the token
	userID, ok := tokenObj.Claims.(jwt.MapClaims)["ID"].(string)
	if !ok {
		return nil, errors.New("GetUserIDFromToken error: type conversion in claims")
	}
	for _, u := range users {
		if string(u.ID) == userID {
			return &u, nil
		}
	}
	return nil, errors.New("No user with ID " + string(userID))
}

func GenerateToken(user *User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID": user.ID,
	})
	tokenString, errToken := token.SignedString([]byte("verysecret"))
	if errToken != nil {
		return "", errToken
	}
	return tokenString, nil
}

func getUser(email, password string) (User, error) {
	for _, u := range users {
		if u.Email == email {
			errCompare := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
			if errCompare != nil {
				return User{}, errCompare
			}
			return u, nil
		}
	}
	return User{}, errors.New("No user with email " + email)
}

func NewAuth(args AuthPayload) (*AuthResolver, error) {
	return &AuthResolver{AuthPayload: args}, nil
}

func (r *AuthResolver) Token() *string {
	return r.AuthPayload.Token
}

func (r *AuthResolver) User(ctx context.Context) (*UserResolver, error) {
	return NewUser(ctx, NewUserArgs{ID: r.AuthPayload.User.ID})
}
