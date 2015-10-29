package go_utils

import "errors"

var (
	ErrInvalidUrl         = errors.New("Invalid url")
	ErrInvalidMail        = errors.New("Invalid e-mail address")
	ErrInvalidMongoIdHash = errors.New("Invalid mongo id hash")
)
