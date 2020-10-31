package resolvers

import (
	"context"
	"errors"
	"github.com/leggettc18/hackernews-clone-api/db"
	"github.com/leggettc18/hackernews-clone-api/model"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type RootResolver struct {
	DB *db.DB
}

func NewRoot(db *db.DB) (*RootResolver, error) {
	return &RootResolver{DB: db}, nil
}

type LinkQueryArgs struct {
	ID uint
}

func (r RootResolver) Link(args LinkQueryArgs) (*LinkResolver, error) {
	link, err := r.DB.GetLinkById(args.ID)
	if err != nil {
		return nil, err
	}
	linkResolver := LinkResolver{
		DB:   r.DB,
		Link: link,
	}

	return &linkResolver, nil
}

type PostArgs struct {
	Description string
	Url         string
}

func (r *RootResolver) Post(ctx context.Context, args PostArgs) (*LinkResolver, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return &LinkResolver{}, errors.New("post: no key 'token' in context")
	}
	author, errAuthor := r.DB.GetUserFromToken(token)
	if errAuthor != nil {
		return &LinkResolver{}, errAuthor
	}
	newLink := model.Link{
		CreatedAt:   time.Now(),
		Description: args.Description,
		Url:         args.Url,
		PostedBy:    *author,
		Votes:       []model.Vote{},
	}

	if err := r.DB.CreateLink(&newLink); err != nil {
		return nil, err
	}
	return &LinkResolver{DB: r.DB, Link: newLink}, nil
}

type LinksQueryArgs struct {
	Or  *[]string
	And *[]string
}

func (r RootResolver) Links(args LinksQueryArgs) (*[]*LinkResolver, error) {
	return NewLinks(NewLinksArgs{Or: args.Or, And: args.And})
}

type SignupArgs struct {
	Email    string
	Password string
	Name     string
}

type UpvoteArgs struct {
	LinkID uint
}

func (r *RootResolver) Upvote(ctx context.Context, args UpvoteArgs) (*VoteResolver, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return &VoteResolver{}, errors.New("post: no key 'token' in context")
	}
	voter, errVoter := r.DB.GetUserFromToken(token)
	if errVoter != nil {
		return &VoteResolver{}, errVoter
	}
	link, err := r.DB.GetLinkById(args.LinkID)
	if err != nil {
		return nil, err
	}
	vote := model.Vote{Link: *link, User: *voter}
	return &VoteResolver{DB: r.DB, Vote: vote}, nil
}

func (r *RootResolver) Signup(args SignupArgs) (*AuthResolver, error) {
	passwordHash, errHash := bcrypt.GenerateFromPassword(
		[]byte(args.Password),
		bcrypt.DefaultCost,
	)
	if errHash != nil {
		return nil, errHash
	}

	newUser := model.User{
		Email:          args.Email,
		HashedPassword: passwordHash,
		Name:           args.Name,
	}

	if err := r.DB.CreateUser(&newUser); err != nil {
		return nil, err
	}

	token, errToken := GenerateToken(&newUser)

	if errToken != nil {
		return nil, errToken
	}

	payload := AuthPayload{
		Token: &token,
		User:  &newUser,
	}

	return &AuthResolver{payload}, nil
}

type LoginArgs struct {
	Email    string
	Password []byte
}

func (r *RootResolver) Login(args LoginArgs) (*AuthResolver, error) {
	user, errUser := r.DB.GetUserByEmail(args.Email)
	if errUser != nil {
		return nil, errUser
	}
	model.ComparePasswordHash(user.HashedPassword, args.Password)

	token, errToken := GenerateToken(user)
	if errToken != nil {
		return nil, errToken
	}
	payload := AuthPayload{
		Token: &token,
		User:  user,
	}
	return &AuthResolver{payload}, nil
}
