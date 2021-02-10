package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/leggettc18/hackernews-clone-api/model"
	"golang.org/x/crypto/bcrypt"
	"time"
)

var (
	passwordHash, _ = bcrypt.GenerateFromPassword(
		[]byte("password"),
		bcrypt.DefaultCost,
	)
	longForm      = "Jan 2, 2006 at 3:04pm (MST)"
	sampleTime, _ = time.Parse(longForm, "Feb 3, 2013 at 7:54pm (PST)")
	users         = []model.User{
		{
			ID:             0,
			Name:           "Christopher Leggett",
			Email:          "chris@leggett.dev",
			HashedPassword: passwordHash,
		},
	}
	links = []model.Link{
		{
			//ID:          0,
			CreatedAt:   sampleTime,
			Url:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PosterID:    0,
		},
		{
			//ID:          1,
			CreatedAt:   sampleTime,
			Url:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PosterID:    0,
		},
		{
			//ID:          2,
			CreatedAt:   sampleTime,
			Url:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PosterID:    0,
		},
		{
			//ID:          3,
			CreatedAt:   sampleTime,
			Url:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PosterID:    0,
		},
		{
			//ID:          4,
			CreatedAt:   sampleTime,
			Url:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PosterID:    0,
		},
		{
			//ID:          5,
			CreatedAt:   sampleTime,
			Url:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PosterID:    0,
		},
		{
			//ID:          6,
			CreatedAt:   sampleTime,
			Url:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PosterID:    0,
		},
		{
			//ID:          7,
			CreatedAt:   sampleTime,
			Url:         "www.howtographql.com",
			Description: "Fullstack tutorial for Graphql",
			PosterID:    0,
		},
	}
	votes = []model.Vote{
		{
			ID:     0,
			UserID: 1,
			LinkID: 1,
		},
	}
)

type DB struct {
	*gorm.DB
}

//NewDB returns a new DB connection.
func NewDB(path string) (*DB, error) {
	// connect to the example db, create it if it doesn't exist.
	db, err := gorm.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	// drop database tables and recreate them fresh
	db.DropTableIfExists(&model.User{}, &model.Link{}, &model.Vote{})
	db.AutoMigrate(&model.User{}, &model.Link{}, &model.Vote{})

	// Insert test data
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return nil, err
		}
	}

	for _, link := range links {
		if err := db.Create(&link).Error; err != nil {
			return nil, err
		}
	}

	for _, vote := range votes {
		if err := db.Create(&vote).Error; err != nil {
			return nil, err
		}
	}

	return &DB{db}, nil
}
