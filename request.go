package go_utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// We can't use jwt.ParseFromRequest() because it calls ParseMultipartForm() and
// it will break MultipartReader() which is important for file upload handling.
func ParseFromRequest(req *http.Request, keyFunc jwt.Keyfunc) (token *jwt.Token, err error) {
	// Look for an Authorization header
	if ah := req.Header.Get("Authorization"); ah != "" {
		// Should be a bearer token
		if len(ah) > 6 && strings.ToUpper(ah[0:6]) == "BEARER" {
			return jwt.Parse(ah[7:], keyFunc)
		}
	}

	return nil, jwt.ErrNoTokenInRequest
}

// ReadAndRewind Reads an io.ReadCloser into an io.Reader and replaces the original source with the buffered source.
// This is required as some readers cannot be Rewind(), for example the http.Request.Body.
func ReadAndRewind(readCloser *io.ReadCloser) (result io.Reader, err error) {
	var content []byte

	// Read the content
	if readCloser != nil {
		content, err = ioutil.ReadAll((*readCloser))
	}

	if err != nil {
		return
	}

	// Restore the io.ReadCloser to its original state
	(*readCloser) = ioutil.NopCloser(bytes.NewBuffer(content))

	return bytes.NewReader(content), err
}
