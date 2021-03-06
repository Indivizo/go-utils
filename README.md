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
