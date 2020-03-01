package model

import (
	"time"

	"cloud.google.com/go/datastore"
)

type User struct {
	ID          *datastore.Key `datastore:"__key__"`
	FirstName   string         `datastore:"first_name"`
	LastName    string         `datastore:"last_name"`
	DateOfBirth time.Time      `datastore:"date_of_birth"`
}
