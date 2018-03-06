package route

import (
  "net/http"

  utils "github.com/Indivizo/go-utils"

  "github.com/julienschmidt/httprouter"
)

const (
  _ = iota
  ErrRouteNotFound
)

type Error struct {
  Type int
  Data interface{}
}

// Error implements error interface.
func (err Error) Error() string {
  switch err.Type {
  case ErrRouteNotFound:
    return "Route not found"

  default:
    return "Unsupported error type"
  }
}

type Route struct {
  Method   string
  Path     string
  Callback httprouter.Handle
  Public   bool
}

type Routes []Route

// GetHttpRouter returns with the http router.
func (routes Routes) GetHttpRouter() *httprouter.Router {
  router := httprouter.New()

  for _, route := range routes {
    switch route.Method {
    case "POST":
      router.POST(route.Path, route.Callback)
    case "GET":
      router.GET(route.Path, route.Callback)
    case "PUT":
      router.PUT(route.Path, route.Callback)
    case "PATCH":
      router.PATCH(route.Path, route.Callback)
    case "DELETE":
      router.DELETE(route.Path, route.Callback)
    }
  }

  return router
}

// GetRouteFromRequest searches for the route from the http request.
func (routes Routes) GetRouteFromRequest(request *http.Request) (*Route, error) {
  var err error
  var bestRoute Route
  var longestMatch int
  for _, route := range routes {
    if route.Method != request.Method {
      continue
    }
    segmentLength := utils.GetMatchingPrefixLength(request.URL.Path, route.Path)
    if segmentLength > longestMatch {
      longestMatch = segmentLength
      bestRoute = route
    }
  }

  if &bestRoute == nil {
    err = Error{Type: ErrRouteNotFound}
    return nil, err
  }

  return &bestRoute, nil
}
