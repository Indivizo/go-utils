package authenticate

import (
  "bytes"
  "encoding/base64"
  "net/http"
  "strings"

  "github.com/Indivizo/go-utils/route"

  log "github.com/Sirupsen/logrus"
  "github.com/gorilla/context"
)

const ContextRequesterApp = "requester_app"

const (
  _ = iota
  ErrApplicationNotFound
  ErrUnsupportedAuthScheme
  ErrWrongAuthToken
  ErrAtuhFail
)

type Error struct {
  Type int
  Data interface{}
}

// Error implements error interface.
func (err Error) Error() string {
  switch err.Type {
  case ErrApplicationNotFound:
    return "Application not found"

  case ErrUnsupportedAuthScheme:
    return "Unsupported authentication scheme"

  case ErrWrongAuthToken:
    return "Wrong authentication token"

  case ErrAtuhFail:
    return "Authentication failed"

  default:
    return "Unsupported error type"
  }
}

type Application struct {
  ID  string
  Key string
}

type Applications []Application

// Push appends an application to Applications.
func (apps *Applications) Push(app Application) {
  (*apps) = append((*apps), app)
}

// GetAppByID searches for on application and return with it.
func (apps Applications) GetAppByID(ID string) (*Application, error) {
  for _, app := range apps {
    if app.ID == ID {
      return &app, nil
    }
  }

  return nil, Error{Type: ErrApplicationNotFound}
}

type Authentication struct {
  Handler     http.Handler
  Routes      route.Routes
  EnabledApps Applications
}

// ServeHTTP is a middleware between router and handler to validate the login token.
func (a Authentication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  // Disable preflight request for AJAX callbacks because it will break the
  // normal (second) request, which contains the correct header.
  if r.Method == "OPTIONS" {
    return
  }

  servedRoute, err := a.Routes.GetRouteFromRequest(r)
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  if servedRoute.Public {
    a.Handler.ServeHTTP(w, r)
  } else {
    app, err := a.AuthenticateFromRequest(r)
    if err != nil {
      http.Error(w, err.Error(), http.StatusBadRequest)
      return
    }
    context.Set(r, ContextRequesterApp, app)
    a.Handler.ServeHTTP(w, r)
    return
  }
}

// AuthenticateFromRequest authenticates the application form the request.
func (a Authentication) AuthenticateFromRequest(req *http.Request) (app *Application, err error) {
  // Validate header token.
  const basicScheme string = "BASIC "
  if ah := req.Header.Get("Authorization"); ah != "" {
    if !strings.HasPrefix(strings.ToUpper(ah), basicScheme) {
      err = Error{Type: ErrUnsupportedAuthScheme}
      return
    }
    var str []byte
    str, err = base64.StdEncoding.DecodeString(ah[len(basicScheme):])
    if err != nil {
      log.WithField("error", err).Warn("Decoding authentication token")
      return
    }
    creds := bytes.SplitN(str, []byte(":"), 2)
    if len(creds) != 2 {
      err = Error{Type: ErrWrongAuthToken}
      return
    }
    givenID := creds[0]
    givenKey := creds[1]

    // Validate app id and key pair.
    app, err = a.EnabledApps.GetAppByID(string(givenID))
    if err != nil {
      return nil, err
    }
    if app.Key == string(givenKey) {
      return
    } else {
      err = Error{Type: ErrAtuhFail}
      return nil, err
    }
  }

  err = Error{Type: ErrAtuhFail}
  return
}
