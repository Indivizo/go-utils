package go_utils

import (
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

const (
	RequestRetryDelay         = 30
	RequestRetrySlowDownLimit = 10
	RequestRetryLimit         = 100

	httpTimeoutInSeconds = 60
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

	// Looking for an access token get query parameter.
	if tokenString := req.URL.Query().Get("access_token"); tokenString != "" {
		return jwt.Parse(tokenString, keyFunc)
	}

	return nil, ErrNoTokenInRequest
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

// Request is a structure to store the details of a network request.
type Request struct {
	ID                 string
	Method             string
	URL                Url
	Body               io.Reader
	ExpectedStatusCode int
	Cancel             chan bool
	Client             *http.Client
	Headers            []http.Header
	Queries            url.Values
	readBuffer         *bytes.Reader
}

// SetupDefaultValues sets up default values for the request structure.
func (request *Request) SetupDefaultValues() {
	if request.Method == "" {
		request.Method = "GET"
	}

	if request.ExpectedStatusCode == 0 {
		request.ExpectedStatusCode = http.StatusOK
	}

	timeout := httpTimeoutInSeconds * time.Second
	dial := (&net.Dialer{
		Timeout:   timeout,
		KeepAlive: timeout,
	}).Dial
	transport := &http.Transport{
		Dial:                dial,
		TLSHandshakeTimeout: timeout,
	}

	if request.Client == nil {
		request.Client = &http.Client{
			Transport: transport,
			Timeout:   timeout,
		}
	} else {
		if request.Client.Transport == nil {
			request.Client.Transport = transport
		} else {
			if tr, ok := request.Client.Transport.(*http.Transport); ok {
				if tr.Dial == nil {
					tr.Dial = dial
				}
				if tr.TLSHandshakeTimeout == 0 {
					tr.TLSHandshakeTimeout = timeout
				}
				request.Client.Transport = tr
			}
		}
		if request.Client.Timeout == 0 {
			request.Client.Timeout = timeout
		}
	}

	if request.Cancel == nil {
		request.Cancel = make(chan bool)
	}
}

// GetHttpRequest returns the http.Request object based on the go-utils.Request
func (request *Request) GetHttpRequest() (req *http.Request, err error) {
	if request.ID == "" {
		request.ID = RandomHash(10)
	}
	// Save the body content to the internal buffer, so we can seek back in case we have to resend the body content (if the request fails).
	if request.readBuffer == nil {
		if request.Body != nil {
			request.readBuffer = new(bytes.Reader)
			buf, _ := ioutil.ReadAll(request.Body)
			request.readBuffer = bytes.NewReader(buf)
		} else {
			request.readBuffer = bytes.NewReader(nil)
		}
	}

	// Seek to the beginning.
	if _, err = request.readBuffer.Seek(0, 0); err != nil {
		return
	}

	// Create request
	if req, err = http.NewRequest(request.Method, request.URL.String(), request.readBuffer); err != nil {
		return
	}

	// Add headers.
	if len(request.Headers) > 0 {
		for _, headers := range request.Headers {
			for key, header := range headers {
				for _, value := range header {
					req.Header.Add(key, value)
				}
			}
		}
	}

	// Add queries.
	req.URL.RawQuery = request.Queries.Encode()

	return
}

func SendRequest(request Request) (response *http.Response, err error) {
	request.SetupDefaultValues()

	req, err := request.GetHttpRequest()
	if err != nil {
		err = errors.Wrap(err, "Getting the HTTP request")
		return
	}
	log.WithFields(log.Fields{
		"requestID": request.ID,
		"method":    request.Method,
		"url":       request.URL,
		"request":   request,
	}).Info("Sending a HTTP request...")

	response, err = request.Client.Do(req)
	logFields := log.Fields{
		"requestID": request.ID,
		"request":   request,
		"error":     err,
		"response":  response,
	}
	if err != nil || response.StatusCode != request.ExpectedStatusCode {
		log.WithFields(logFields).Warn("Send request was unsuccessful")
	} else {
		log.WithFields(logFields).Info("Send request was successful")
	}

	return
}

func SendRequestWithRetry(request Request) chan *http.Response {
	request.SetupDefaultValues()

	response := make(chan *http.Response)

	go func() {
		if resp, err := SendRequest(request); err != nil || resp.StatusCode != request.ExpectedStatusCode {
			quit := false
			tries := 1
			delay := RequestRetryDelay
			for {
				select {
				case <-request.Cancel:
					log.WithFields(log.Fields{
						"requestID": request.ID,
						"tries":     tries,
					}).Warn("Request canceled")
					close(response)
					quit = true
					break

				default:
					time.Sleep(time.Second * time.Duration(delay))

					if resp, err = SendRequest(request); err != nil || resp.StatusCode != request.ExpectedStatusCode {
						log.WithFields(log.Fields{
							"requestID": request.ID,
							"tries":     tries,
						}).Warn("Request failed with retry")
					} else {
						response <- resp
						quit = true
						break
					}

					// Set delay.
					if tries%RequestRetrySlowDownLimit == 0 {
						delay = delay * 2
					}
					// Quit after x tries.
					if tries == RequestRetryLimit {
						log.WithFields(log.Fields{
							"requestID": request.ID,
							"tries":     tries,
						}).Warn("Request failed. Stop retrying.")
						close(response)
						quit = true
						break
					}

					tries++
				}

				// Break from the for loop.
				if quit {
					break
				}
			}
		} else {
			response <- resp
		}
	}()

	return response
}
