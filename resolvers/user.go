package resolvers

import (
	"context"
	"errors"
	"github.com/graph-gophers/graphql-go"
)

type User struct {
	ID       graphql.ID
	Name     string
	Email    string
	Password string
	Links    []Link
	Votes    []Vote
}

type UserResolver struct {
	User User
}

type NewUserArgs struct {
	ID graphql.ID
}

func NewUser(ctx context.Context, args NewUserArgs) (*UserResolver, error) {
	for _, user := range users {
		if user.ID == args.ID {
			return &UserResolver{User: user}, nil
		}
	}
	return nil, errors.New("User not found")
}
