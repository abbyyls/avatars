package main

import (
    "encoding/json"
	"fmt"
	"gopkg.in/mgo.v2"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
)

const MaxFileSize = 10 * 1024 * 1024 // 10 MB


type httperror interface {
    error
    Status() int  // http status code
}

// HttpError is an httperror implementation that includes a http status code and message.
type HttpError struct {
	status int
	message string
}

func (e HttpError) Error() string {
	return fmt.Sprint(e.message)
}

func (e HttpError) Status() int {
	return e.status
}

// Check whether the file size is acceptable.
func checkFileSize(r io.Reader) httperror {
	data, err := ioutil.ReadAll(io.LimitReader(r, MaxFileSize + 1))
	if err != nil {
		return HttpError{http.StatusInternalServerError, err.Error()}
	}
	if len(data) > MaxFileSize {
		return HttpError{http.StatusRequestEntityTooLarge, `file is too large`}
	}
	return nil
}

// Check whether a string slice contains a certain value.
func contains(slice []string, value string) bool {
    for _, item := range slice {
		if item == value {
			return true
		}
	}
    return false
}

// Get content type of GridFS-file if set. Otherwise guess by extension.
// If it cannot determine a more specific one,
// it returns "application/octet-stream".
func getMimeType(file *mgo.GridFile) string {
	filetype := file.ContentType()
	if filetype == "" {
		filetype := mime.TypeByExtension(filepath.Ext(file.Name()))
		if filetype == "" {
			return "application/octet-stream"
		}
		return filetype
	}
	return filetype
}

// Write JSON-response with given status code and message.
// JSON struct: {"msg": "some message"}
func JsonResponseMsg(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"msg": msg}); err != nil {
		panic(err)
	}
    return
}

// Write JSON-response with given status code and struct object.
func JsonResponseFromStruct(w http.ResponseWriter, status int, avatar *Avatar) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	jsonString, err := json.Marshal(avatar)
	if err != nil {
		panic(err)
	}
	w.Write(jsonString)
    return
}
