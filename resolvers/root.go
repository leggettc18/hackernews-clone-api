package resolvers

import (
	"context"
	"errors"
	"fmt"
	"github.com/graph-gophers/graphql-go"
	"github.com/leggettc18/hackernews-clone-api/db"
	"github.com/leggettc18/hackernews-clone-api/model"
	"golang.org/x/crypto/bcrypt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"
)

type RootResolver struct {
	DB                *db.DB
	NewLinkEvents     chan *NewLinkEvent
	NewLinkSubscriber chan *NewLinkSubscriber
	NewVoteEvents     chan *NewVoteEvent
	NewVoteSubscriber chan *NewVoteSubscriber
}

type NewLinkSubscriber struct {
	stop   <-chan struct{}
	events chan<- *NewLinkEvent
}

type NewLinkEvent struct {
	EventID string
	Link    *LinkResolver
}

func (r *NewLinkEvent) NewLink() *LinkResolver {
	return r.Link
}

func (r *NewLinkEvent) ID() string {
	return r.EventID
}

type NewVoteEvent struct {
	EventID string
	Vote    *VoteResolver
}

func (r *NewVoteEvent) NewVote() *VoteResolver {
	return r.Vote
}

func (r *NewVoteEvent) ID() string {
	return r.EventID
}

type NewVoteSubscriber struct {
	stop   <-chan struct{}
	events chan<- *NewVoteEvent
}

func NewRoot(db *db.DB) (*RootResolver, error) {
	r := &RootResolver{
		DB:                db,
		NewLinkEvents:     make(chan *NewLinkEvent),
		NewLinkSubscriber: make(chan *NewLinkSubscriber),
		NewVoteEvents:     make(chan *NewVoteEvent),
		NewVoteSubscriber: make(chan *NewVoteSubscriber),
	}

	go r.broadcastNewLink()
	go r.broadcastNewVote()

	return r, nil
}

func (r *RootResolver) broadcastNewLink() {
	subscribers := map[string]*NewLinkSubscriber{}
	unsubscribe := make(chan string)

	for {
		select {
		case id := <-unsubscribe:
			delete(subscribers, id)
		case s := <-r.NewLinkSubscriber:
			subscribers[randomID()] = s
		case e := <-r.NewLinkEvents:
			for id, s := range subscribers {
				go func(id string, s *NewLinkSubscriber) {
					select {
					case <-s.stop:
						unsubscribe <- id
						return
					default:
					}

					select {
					case <-s.stop:
						unsubscribe <- id
					case s.events <- e:
					case <-time.After(time.Second):
					}
				}(id, s)
			}
		}
	}
}

func (r *RootResolver) broadcastNewVote() {
	subscribers := map[string]*NewVoteSubscriber{}
	unsubscribe := make(chan string)

	for {
		select {
		case id := <-unsubscribe:
			delete(subscribers, id)
		case s := <-r.NewVoteSubscriber:
			subscribers[randomID()] = s
		case e := <-r.NewVoteEvents:
			for id, s := range subscribers {
				go func(id string, s *NewVoteSubscriber) {
					select {
					case <-s.stop:
						unsubscribe <- id
						return
					default:
					}

					select {
					case <-s.stop:
						unsubscribe <- id
					case s.events <- e:
					case <-time.After(time.Second):
					}
				}(id, s)
			}
		}
	}
}

func randomID() string {
	var letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, 16)
	for i := range b {
		b[i] = letter[rand.Intn(len(letter))]
	}
	return string(b)
}

func (r *RootResolver) NewLink(ctx context.Context) (<-chan *NewLinkEvent, error) {
	fmt.Println("subscribing to new links")
	c := make(chan *NewLinkEvent)
	r.NewLinkSubscriber <- &NewLinkSubscriber{events: c, stop: ctx.Done()}
	return c, nil
}

func (r *RootResolver) NewVote(ctx context.Context) (<-chan *NewVoteEvent, error) {
	fmt.Println("subscribing to new links")
	c := make(chan *NewVoteEvent)
	r.NewVoteSubscriber <- &NewVoteSubscriber{events: c, stop: ctx.Done()}
	return c, nil
}

type LinkQueryArgs struct {
	ID graphql.ID
}

func (r RootResolver) Link(args LinkQueryArgs) (*LinkResolver, error) {
	id, err := getUintFromGraphqlId(args.ID)
	if err != nil {
		return nil, err
	}
	link, err := r.DB.GetLinkById(id)
	if err != nil {
		return nil, err
	}
	linkResolver := LinkResolver{
		DB:   r.DB,
		Link: *link,
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
		PosterID:    author.ID,
		Votes:       []model.Vote{},
	}

	if err := r.DB.CreateLink(&newLink); err != nil {
		return nil, err
	}
	linkResolver := &LinkResolver{DB: r.DB, Link: newLink}

	select {
	case r.NewLinkEvents <- &NewLinkEvent{Link: linkResolver, EventID: randomID()}:
		// values are being read from r.Events
		fmt.Println("r.NewLinks: inserted link")
	default:
		//no subscribers, link not in channel
		fmt.Println("r.NewLinks: link created, not inserted")
	}

	return linkResolver, nil
}

type LinksQueryArgs struct {
	Or      *[]string
	And     *[]string
	First   *float64
	Skip    *float64
	OrderBy *string
}

func (r RootResolver) LinksMeta() (*MetaResolver, error) {
	var links []model.Link
	result := r.DB.Find(&links)
	err := result.Error
	if err != nil {
		return nil, err
	}
	return &MetaResolver{int32(result.RowsAffected)}, nil
}

func (r RootResolver) Links(args LinksQueryArgs) (*[]*LinkResolver, error) {
	var (
		links   []model.Link
		results []model.Link
	)
	if err := r.DB.Find(&links).Error; err != nil {
		return nil, err
	}
	if args.And != nil {
		for _, link := range links {
			hasAllTerms := true
			for _, term := range *args.And {
				if hasAllTerms == false {
					break
				}
				if strings.Contains(link.Description, term) || strings.Contains(link.Url, term) {
					hasAllTerms = true
				} else {
					hasAllTerms = false
				}
			}
			if hasAllTerms == true {
				results = append(results, link)
			}
		}
	} else if args.Or != nil {
		for _, link := range links {
			for _, term := range *args.Or {
				if strings.Contains(link.Description, term) || strings.Contains(link.Url, term) {
					results = append(results, link)
					break
				}
			}
		}
	} else {
		results = links
	}

	if args.Skip != nil {
		if args.First != nil && (len(results) > int(*args.First+*args.Skip)) {
			results = results[int(*args.Skip):int(*args.First+*args.Skip)]
		} else {
			results = results[int(*args.Skip):]
		}
	}

	if args.OrderBy != nil {
		if *args.OrderBy == "createdAt_DESC" {
			sort.Slice(results, func(i, j int) bool {
				return results[i].CreatedAt.Before(results[j].CreatedAt)
			})
		}
	}
	var resolvers []*LinkResolver
	for _, link := range results {
		resolvers = append(resolvers, &LinkResolver{r.DB, link})
	}
	return &resolvers, nil
}

type SignupArgs struct {
	Email    string
	Password string
	Name     string
}

type UpvoteArgs struct {
	LinkID graphql.ID
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
	id, err := getUintFromGraphqlId(args.LinkID)
	if err != nil {
		return nil, err
	}
	link, err := r.DB.GetLinkById(id)
	if err != nil {
		return nil, err
	}
	vote := model.Vote{LinkID: link.ID, UserID: voter.ID}
	if err := r.DB.CreateVote(&vote); err != nil {
		return nil, err
	}
	voteResolver := &VoteResolver{DB: r.DB, Vote: vote}
	select {
	case r.NewVoteEvents <- &NewVoteEvent{Vote: voteResolver, EventID: randomID()}:
		// values are being read from r.Events
		fmt.Println("r.NewVotes: inserted vote")
	default:
		//no subscribers, link not in channel
		fmt.Println("r.NewVotes: vote created, not inserted")
	}
	return voteResolver, nil
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

	return &AuthResolver{r.DB, payload}, nil
}

type LoginArgs struct {
	Email    string
	Password string
}

func (r *RootResolver) Login(args LoginArgs) (*AuthResolver, error) {
	user, errUser := r.DB.GetUserByEmail(args.Email)
	if errUser != nil {
		return nil, errUser
	}
	model.ComparePasswordHash(user.HashedPassword, []byte(args.Password))

	token, errToken := GenerateToken(user)
	if errToken != nil {
		return nil, errToken
	}
	payload := AuthPayload{
		Token: &token,
		User:  user,
	}
	return &AuthResolver{r.DB, payload}, nil
}

//Helpers
func getUintFromGraphqlId(gqlid graphql.ID) (uint, error) {
	id, err := strconv.ParseUint(string(gqlid), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
