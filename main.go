package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

var (
	opts            = []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	passwordHash, _ = bcrypt.GenerateFromPassword(
		[]byte("password"),
		bcrypt.DefaultCost,
	)
	longForm      = "Jan 2, 2006 at 3:04pm (MST)"
	sampleTime, _ = time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
	users         = []User{
		{
			ID:       "0",
			Name:     "Christopher Leggett",
			Email:    "chris@leggett.dev",
			Password: string(passwordHash),
		},
	}
	links = []Link{
		{
			ID:          "0",
			CreatedAt:   graphql.Time{Time: sampleTime},
			URL:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PostedBy:    users[0],
		},
	}
	votes = []Vote{
		{
			ID:   "0",
			User: users[0],
			Link: links[0],
		},
	}
)

var (
	addr              = ":8081"
	readHeaderTimeout = 1 * time.Second
	writeTimeout      = 10 * time.Second
	idleTimeout       = 90 * time.Second
	maxHeaderBytes    = http.DefaultMaxHeaderBytes
)

type User struct {
	ID       graphql.ID
	Name     string
	Email    string
	Password string
	Links    []Link
	Votes    []Vote
}

type AuthPayload struct {
	Token *string
	User  *User
}

type RootResolver struct{}

type Link struct {
	ID          graphql.ID
	CreatedAt   graphql.Time
	Description string
	URL         string
	PostedBy    User
	Votes       []Vote
}

type Vote struct {
	ID   graphql.ID
	User User
	Link Link
}

func (r *RootResolver) Info() (string, error) {
	return "this is a thing", nil
}

func (r *RootResolver) Feed() ([]Link, error) {
	var processedLinks = []Link{}
	for index, link := range links {
		processedLinks = append(processedLinks, link)
		for _, vote := range votes {
			if vote.Link.ID == processedLinks[index].ID {
				processedLinks[index].Votes = append(processedLinks[index].Votes, vote)
			}
		}
	}
	return processedLinks, nil
}

func (r *RootResolver) Link(args struct {
	ID graphql.ID
}) (Link, error) {
	for _, link := range links {
		if link.ID == args.ID {
			for _, vote := range votes {
				if vote.Link.ID == link.ID {
					link.Votes = append(link.Votes, vote)
				}
			}
			return link, nil
		}
	}
	return Link{
		ID:          "",
		Description: "",
		URL:         "",
	}, errors.New("ID not found")
}

func (r *RootResolver) Post(
	ctx context.Context,
	args struct {
		Description string
		URL         string
	}) (Link, error) {
	token, ok := ctx.Value("token").(string)
	if !ok {
		return Link{}, errors.New("Post: no key 'token' in context")
	}
	author, errAuthor := GetUserFromToken(token)
	if errAuthor != nil {
		return Link{}, errAuthor
	}
	newLink := Link{
		ID:          graphql.ID(fmt.Sprint(len(links))),
		Description: args.Description,
		URL:         args.URL,
		PostedBy:    *author,
	}

	links = append(links, newLink)
	return newLink, nil
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

func (r *RootResolver) Signup(args struct {
	Email    string
	Password string
	Name     string
}) (*AuthPayload, error) {
	passwordHash, errHash := bcrypt.GenerateFromPassword(
		[]byte(args.Password),
		bcrypt.DefaultCost,
	)
	if errHash != nil {
		return nil, errHash
	}

	newUser := User{
		ID:       graphql.ID(fmt.Sprint(len(users))),
		Email:    args.Email,
		Password: string(passwordHash),
		Name:     args.Name,
	}

	users = append(users, newUser)

	token, errToken := GenerateToken(&newUser)

	if errToken != nil {
		return nil, errToken
	}

	return &AuthPayload{
		Token: &token,
		User:  &newUser,
	}, nil
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

func (r *RootResolver) Login(args struct {
	Email    string
	Password string
}) (*AuthPayload, error) {
	user, errUser := getUser(args.Email, args.Password)
	if errUser != nil {
		return nil, errUser
	}

	token, errToken := GenerateToken(&user)
	if errToken != nil {
		return nil, errToken
	}
	return &AuthPayload{
		Token: &token,
		User:  &user,
	}, nil
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

// Reads and parses the schema from file.
// Associates root resolver. Panics if can't read.
func parseSchema(path string, resolver interface{}) *graphql.Schema {
	bstr, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	schemaString := string(bstr)
	parsedSchema, err := graphql.ParseSchema(
		schemaString,
		resolver,
		opts...,
	)
	if err != nil {
		panic(err)
	}
	return parsedSchema
}

func main() {
	mux := http.NewServeMux()

	gqlHandler := &relay.Handler{
		Schema: parseSchema("./schema.graphql", &RootResolver{}),
	}

	mux.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		token := strings.ReplaceAll(r.Header.Get("Authorization"), "Bearer ", "")
		ctx := context.WithValue(context.Background(), "token", token)
		gqlHandler.ServeHTTP(w, r.WithContext(ctx))
	})

	// necessary CORS options. Should not be used in production
	// AllowedOrigins should be more specific than * and the
	// AllowOriginFunc should not be present. This code is not
	// written for production or for a CORS tutorial so this is fine
	// for its purpose.
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowOriginFunc:  func(origin string) bool { return true },
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	}).Handler(mux)

	s := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}

	// Begin listeing for requests.
	log.Printf("Listening for requests on %s", s.Addr)

	if err := s.ListenAndServe(); err != nil {
		log.Println("server.ListenAndServe:", err)
	}
}
