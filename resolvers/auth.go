package resolvers

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/leggettc18/hackernews-clone-api/db"
	"github.com/leggettc18/hackernews-clone-api/model"
)

type AuthResolver struct {
	DB          *db.DB
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

func (r *AuthResolver) Token() *string {
	return r.AuthPayload.Token
}

func (r *AuthResolver) User() *UserResolver {
	return &UserResolver{r.DB, *r.AuthPayload.User}
}
