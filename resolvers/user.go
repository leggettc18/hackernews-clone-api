package resolvers

import (
	goErrors "errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/leggettc18/hackernews-clone-api/db"
	"github.com/leggettc18/hackernews-clone-api/errors"
	"github.com/leggettc18/hackernews-clone-api/model"
)

type UserResolver struct {
	DB   *db.DB
	User model.User
}

type NewUserArgs struct {
	ID uint
}

func NewUser(args NewUserArgs) (*UserResolver, error) {
	for _, user := range users {
		if user.ID == args.ID {
			return &UserResolver{User: user}, nil
		}
	}
	return nil, goErrors.New("user not found")
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

func (r *UserResolver) Links() (*[]*LinkResolver, error) {
	var (
		resolvers = make([]*LinkResolver, len(r.User.Links))
		errs      errors.Errors
	)
	for index, link := range links {
		if link.PostedBy.ID == r.User.ID {
			resolver, err := NewLink(NewLinkArgs{ID: link.ID})
			if err != nil {
				errs = append(errs, errors.WithIndex(err, index))
			}
			resolvers = append(resolvers, resolver)
		}
	}
	if errs != nil {
		return &resolvers, errs.Err()
	}
	return &resolvers, nil
}

func (r *UserResolver) Votes() (*[]*VoteResolver, error) {
	var (
		resolvers = make([]*VoteResolver, len(r.User.Votes))
		errs      errors.Errors
	)
	for index, vote := range votes {
		if vote.User.ID == r.User.ID {
			resolver, err := NewVote(NewVoteArgs{ID: vote.ID})
			if err != nil {
				errs = append(errs, errors.WithIndex(err, index))
			} else {
				resolvers = append(resolvers, resolver)
			}
		}
	}
	if errs != nil {
		return &resolvers, errs.Err()
	}
	return &resolvers, nil
}
