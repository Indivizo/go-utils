package go_utils

import (
	"github.com/asaskevich/govalidator"
)

type Url string

func (url Url) String() string {
	return (string)(url)
}

func (url Url) Validate() error {
	if !govalidator.IsRequestURL(url.String()) {
		return ErrInvalidUrl
	}
	return nil
}

type Email string

func (email Email) String() string {
	return (string)(email)
}

func (email Email) Validate() error {
	if !govalidator.IsEmail(email.String()) {
		return ErrInvalidMail
	}
	return nil
}
