package go_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUrl(t *testing.T) {
	var url Url
	var err error

	url = "example"
	err = url.Validate()
	assert.Equal(t, ErrInvalidUrl, err)

	url = "example.com"
	err = url.Validate()
	assert.Equal(t, ErrInvalidUrl, err)

	url = "www.example.com"
	err = url.Validate()
	assert.Equal(t, ErrInvalidUrl, err)

	url = "http://example.com"
	err = url.Validate()
	assert.Nil(t, err)
}

func TestEmail(t *testing.T) {
	var email Email
	var err error

	email = "example"
	err = email.Validate()
	assert.Equal(t, ErrInvalidMail, err)

	email = "example.com"
	err = email.Validate()
	assert.Equal(t, ErrInvalidMail, err)

	email = "www.example.com"
	err = email.Validate()
	assert.Equal(t, ErrInvalidMail, err)

	email = "test@example.com"
	err = email.Validate()
	assert.Nil(t, err)

	email = "test+test1@example.com"
	err = email.Validate()
	assert.Nil(t, err)
}
