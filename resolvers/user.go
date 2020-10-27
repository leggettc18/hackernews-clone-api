package resolvers

import (
	"context"
	goErrors "errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/leggettc18/hackernews-clone-api/errors"
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
	return nil, goErrors.New("User not found")
}

func (r *UserResolver) ID() graphql.ID {
	return r.User.ID
}

func (r *UserResolver) Name() string {
	return r.User.Name
}

func (r *UserResolver) Email() string {
	return r.User.Email
}

func (r *UserResolver) Links(ctx context.Context) (*[]*LinkResolver, error) {
	var (
		resolvers = make([]*LinkResolver, len(r.User.Links))
		errs      errors.Errors
	)
	for index, link := range links {
		if link.PostedBy.ID == r.User.ID {
			resolver, err := NewLink(ctx, NewLinkArgs{ID: link.ID})
			if err != nil {
				errs = append(errs, errors.WithIndex(err, index))
			}
			resolvers = append(resolvers, resolver)
		}
	}
	return &resolvers, errs.Err()
}

func (r *UserResolver) Votes(ctx context.Context) (*[]*VoteResolver, error) {
	var (
		resolvers = make([]*VoteResolver, len(r.User.Votes))
		errs      errors.Errors
	)
	for index, vote := range votes {
		if vote.User.ID == r.User.ID {
			resolver, err := NewVote(ctx, NewVoteArgs{ID: vote.ID})
			if err != nil {
				errs = append(errs, errors.WithIndex(err, index))
			} else {
				resolvers = append(resolvers, resolver)
			}
		}
	}
	return &resolvers, errs.Err()
}
