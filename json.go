package go_utils

import (
	"encoding/json"
	"net/http"
	"strconv"

	"labix.org/v2/mgo"
)

func RenderDataAsJSON(w http.ResponseWriter, data interface{}, err error, httpStatus int, httpErrorStatus int) {
	if err != nil {
		if err == mgo.ErrNotFound {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), httpErrorStatus)
		return
	}

	// Stream JSON to the output.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	encoder := json.NewEncoder(w)
	if err := encoder.Encode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Writes the given data as a json and set the status code.
func WriteJson(w http.ResponseWriter, data []byte, status int) {
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Content-Type", "application/json")
	// Set the response status here to avoid multiple WriteHeader calls.
	w.WriteHeader(status)
	w.Write(data)
}
