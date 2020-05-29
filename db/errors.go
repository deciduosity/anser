package db

import (
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

var errNotFound = errors.New("document not found")

func ResultsNotFound(err error) bool {
	if err == nil {
		return false
	}
	err = errors.Cause(err)
	return err.Error() == "not found" || err == errNotFound || err == mongo.ErrNoDocuments
}
