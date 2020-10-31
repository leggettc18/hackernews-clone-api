package resolvers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/leggettc18/hackernews-clone-api/model"
)

type AuthResolver struct {
	AuthPayload AuthPayload
}

type AuthPayload struct {
	Token *string
	User  *model.User
}

func GenerateToken(user *model.User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"ID": user.ID,
	})
	tokenString, errToken := token.SignedString([]byte("verysecret"))
	if errToken != nil {
		return "", errToken
	}
	return tokenString, nil
}
func NewAuth(args AuthPayload) (*AuthResolver, error) {
	return &AuthResolver{AuthPayload: args}, nil
}

func (r *AuthResolver) Token() *string {
	return r.AuthPayload.Token
}

func (r *AuthResolver) User() (*UserResolver, error) {
	return NewUser(NewUserArgs{ID: r.AuthPayload.User.ID})
}
