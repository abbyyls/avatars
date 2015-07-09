package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
)

const MaxFileSize = 10 * 1024 * 1024 // 10 MB

// Check whether the file size is acceptable.
func checkFileSize(r io.Reader) error {
	data, err := ioutil.ReadAll(io.LimitReader(r, MaxFileSize+1))
	if err != nil {
		return err
	}
	if len(data) > MaxFileSize {
		return errors.New(`CONTENT_LENGTH_TOO_LARGE`)
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

// Get content type of file if set. Otherwise returns "application/octet-stream".
func getFileType(file io.Reader) (reader *bytes.Reader, filetype string, err error) {
	var array []byte
	if array, err = ioutil.ReadAll(file); err != nil {
		return
	}
	filetype = http.DetectContentType(array)
	reader = bytes.NewReader(array)
	return
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
