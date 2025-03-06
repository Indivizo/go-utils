# Go utils
Helper package for our go projects.

## Slice

#### StringInSlice()

## JSON

#### RenderDataAsJSON()
#### WriteJson()

## Request

#### ParseFromRequest()
#### ReadAndRewind()
#### GetMatchingPrefixLength()
#### SendRequest()
This function will try to send the request in the background several times if the response is failes.

**Example usage:**
```go
type MyType struc {
  Field1 string `json:"field1"`
  Field2 string `json:"field2"`
}
myValue := MyType{
  Field1: "Hello",
  Field2: "GO",
}
queryBody, _ := json.Marshal(myValue)
contentReader := bytes.NewBuffer(queryBody)
  go func() {
    cancel := make(chan bool)
    res := utils.SendRequest(utils.Request{
      URL:     "http://example.com",
      Method:  "POST",
      Body:    contentReader,
      Headers: []http.Header{{"Content-Type": []string{"application/json"}}},
    })

    response := <-res
    if response == nil {
      log.Println("request failed")
    } else {
      log.Printf("cool, the reponse is: %v\n", response)
    }
  }()
```
If you want to cancel the request just set it in the `cancel` chanel:
```go
go func() {
    cancel := make(chan bool)
    res := utils.SendRequest(utils.Request{
      URL:     "http://example.com",
      Method:  "POST",
      Body:    contentReader,
      Headers: []http.Header{{"Content-Type": []string{"application/json"}}},
      Cancel:  cancel,
    })
    time.Sleep(time.Second * 5)
    cancel <- true
  }()
```
Or if you want to combine them:
```go
go func() {
    cancel := make(chan bool)
    res := utils.SendRequest(utils.Request{
      URL:     "http://example.com",
      Method:  "POST",
      Body:    contentReader,
      Headers: []http.Header{{"Content-Type": []string{"application/json"}}},
      Cancel:  cancel,
    })
    go func() {
      time.Sleep(time.Second * 5)
      cancel <- true
    }()
    response := <-res
    if response == nil {
      log.Println("this request is failed")
    } else {
      log.Printf("cool, the reponse is: %v\n", response)
    }
  }()
```

## Misc

#### RandomHash()

## Error

### Not Found Error Handling

This package provides tools for consistent handling of "not found" errors. These tools allow services to uniformly address cases when a resource cannot be found.

#### ErrNotFound
A generic error representing a "resource not found" case:
```go
var ErrNotFound = errors.New("resource not found")
```

#### RegisterNotFoundError()
Allows services to register custom "not found" errors specific to their domain (e.g., a database error such as MongoDB's `mongo.ErrNoDocuments`).

**Example 1:**
```go
import (
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
)

// Register the MongoDB "not found" error.
utils.RegisterNotFoundError(mongo.ErrNoDocuments)

// Usage example:
err := mongo.ErrNoDocuments
if utils.IsNotFoundError(err) {
	fmt.Println("This is a 'not found' error")
}
```


**Example 2:**
```go
dbError := errors.New("database entry not found")
utils.RegisterNotFoundError(dbError)
```

#### IsNotFoundError()
Checks whether the given error matches `ErrNotFound` or any of the registered custom "not found" errors.

**Example:**
```go
err := errors.New("database entry not found")
if utils.IsNotFoundError(err) {
    fmt.Println("This is a 'not found' error")
}
```

## Other errors

#### ErrInvalidUrl
#### ErrInvalidMail
#### ErrInvalidMongoIdHash

## Types

#### Url
* String()
* Validate()

#### Email
* String()
* Validate()
