package go_utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
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

// GetMatchingPrefixLength returns the length of the path a pattern could match as a prefix.
// Supports parameters using the ":parameter" or "*parameter" notation.
func GetMatchingPrefixLength(path, pattern string) int {
	// Drop the leading slashes.
	path = strings.Trim(path, "/")
	pattern = strings.Trim(pattern, "/")

	pathSegments := strings.Split(path, "/")
	patternSegments := strings.Split(pattern, "/")

	// If the pattern is longer than the path, we will surely can't match it.
	if len(patternSegments) > len(pathSegments) {
		return 0
	}

	i := 0
	for ; i < len(pathSegments); i++ {
		if i >= len(patternSegments) || // Run out of pattern segments to match
			!(pathSegments[i] == patternSegments[i] || // Either the segments has to match
				strings.HasPrefix(patternSegments[i], ":") || strings.HasPrefix(patternSegments[i], "*")) { // Or the segment is a parameter
			break
		}
	}

	return i
}

func SendRequest(method string, url string, body io.Reader, expectedStatusCode int, extraHeaders ...http.Header) chan *http.Response {
	response := make(chan *http.Response)
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		close(response)
		return response
	}
	if len(extraHeaders) > 0 {
		for _, headers := range extraHeaders {
			for key, header := range headers {
				for _, value := range header {
					req.Header.Add(key, value)
				}
			}
		}
	}

	client := &http.Client{}
	go func() {
		if resp, err := client.Do(req); err != nil || resp.StatusCode != expectedStatusCode {
			log.WithFields(log.Fields{
				"method":   method,
				"url":      url,
				"error":    err,
				"response": resp,
			}).Warn("Send request failed")

			tries := 1
			delay := 30
			for {
				time.Sleep(time.Second * time.Duration(delay))

				if resp, err := client.Do(req); err != nil || resp.StatusCode != expectedStatusCode {
					log.WithFields(log.Fields{
						"try":      tries,
						"method":   method,
						"url":      url,
						"error":    err,
						"response": resp,
					}).Warn("Send request failed in queue")
				} else {
					log.WithFields(log.Fields{
						"method": method,
						"url":    url,
						"tries":  tries,
					}).Info("Send request is successfull in queue")
					response <- resp
					break
				}

				// Set delay.
				if tries%10 == 0 {
					delay = delay * 2
				}
				// Quit after 100 tries.
				if tries == 100 {
					log.WithFields(log.Fields{
						"method": method,
						"url":    url,
						"try":    tries,
					}).Error("Send request is failed. The request is removed from the queue")
					close(response)
					break
				}

				tries++
			}
		} else {
			response <- resp
		}
	}()

	return response
}
