package storage

import "errors"

var (
	ErrURLNotFound = errors.New("URL not found")
	ErrURLExist    = errors.New("URL exists")
)
