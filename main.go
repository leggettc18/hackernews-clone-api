package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

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

	// cors.Default() setup the middleware with default options being
	// all origins accepted with simple methods (GET, POST). See
	// documentation below for more options.
	handler := cors.Default().Handler(mux)

	mux.Handle("/graphql", &relay.Handler{
		Schema: parseSchema("./schema.graphql", &RootResolver{}),
	})

	fmt.Println("serving on 8081")
	log.Fatal(http.ListenAndServe(":8081", handler))
}
