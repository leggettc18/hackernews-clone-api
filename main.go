package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/rs/cors"
	"golang.org/x/crypto/bcrypt"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
)

var (
	opts  = []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	links = []Link{
		{
			ID:          "0",
			URL:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
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

var users []User = []User{}

type User struct {
	ID       graphql.ID
	Name     string
	Email    string
	Password string
	Links    []Link
}

type AuthPayload struct {
	Token *string
	User  *User
}

type RootResolver struct{}

type Link struct {
	ID          graphql.ID
	Description string
	URL         string
	PostedBy    Poster
}

type Poster struct {
	ID   graphql.ID
	Name string
}

func (r *RootResolver) Info() (string, error) {
	return "this is a thing", nil
}

func (r *RootResolver) Feed() ([]Link, error) {
	return links, nil
}

func (r *RootResolver) Link(args struct {
	ID graphql.ID
}) (Link, error) {
	for _, link := range links {
		if link.ID == args.ID {
			return link, nil
		}
	}
	return Link{
		ID:          "",
		Description: "",
		URL:         "",
	}, errors.New("ID not found")
}

func (r *RootResolver) Post(args struct {
	Description string
	URL         string
	PostedBy    string
}) (Link, error) {
	userName, errID := getNameFromID(args.PostedBy)
	if errID != nil {
		return Link{}, errID
	}
	newLink := Link{
		ID:          graphql.ID(fmt.Sprint(len(links))),
		Description: args.Description,
		URL:         args.URL,
		PostedBy: Poster{
			ID:   graphql.ID(fmt.Sprint(args.PostedBy)),
			Name: userName,
		},
	}

	links = append(links, newLink)
	return newLink, nil
}

func getNameFromID(id string) (string, error) {
	for _, user := range users {
		if user.ID == graphql.ID(id) {
			return user.Name, nil
		}
	}
	return "", errors.New("User ID not found")
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

	mux.Handle("/graphql", &relay.Handler{
		Schema: parseSchema("./schema.graphql", &RootResolver{}),
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
