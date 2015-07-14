package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/zenazn/goji/web"
)

type CheckIdMiddlewareSuiteTester struct {
	BaseSuite

	mux *web.Mux
}

// Settings for suite
func (suite *CheckIdMiddlewareSuiteTester) SetupSuite() {
	// INIT test router
	router := func(c web.C, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(c.URLParams["id"]))
	}
	// AND set mux with 'CheckId' middleware
	suite.mux = web.New()
	RouterWithId := web.New()
	RouterWithId.Use(CheckId)
	RouterWithId.Get("/:id", router)
	suite.mux.Handle("/:id", RouterWithId)
}

// Test CheckId middleware with valid id
func (suite *CheckIdMiddlewareSuiteTester) TestValidId() {
	// GIVEN valid id
	id := RandomMD5()

	// WHEN I send new request
	r, err := http.NewRequest("GET", "/"+id, nil)
	if err != nil {
		suite.T().Error(err.Error())
	}
	w := httptest.NewRecorder()
	// AND handle response
	suite.mux.ServeHTTP(w, r)
	// THEN response status code should be 200
	suite.Equal(w.Code, http.StatusOK, fmt.Sprintf("status is %d, should be %d", w.Code, http.StatusOK))
	// AND response body should contain id
	suite.Equal(w.Body.String(), id, fmt.Sprintf("body is %q, should be %q", w.Body.String(), id))
}

// Test CheckId middleware with invalid id length
func (suite *CheckIdMiddlewareSuiteTester) TestInvalidIdLength() {
	// GIVEN invalid id with short length
	id := "7c24e1a"

	// WHEN I send new request
	r, err := http.NewRequest("GET", "/"+id, nil)
	if err != nil {
		suite.T().Error(err.Error())
	}
	w := httptest.NewRecorder()
	// AND handle response
	suite.mux.ServeHTTP(w, r)
	// THEN response status code should be 400
	suite.Equal(w.Code, http.StatusBadRequest, fmt.Sprintf("status is %d, should be %d", w.Code, http.StatusBadRequest))
	// AND response body should contain error message
	decoder := json.NewDecoder(w.Body)
	body := struct {
		Msg string `json:"msg"`
	}{}
	err = decoder.Decode(&body)
	if err != nil {
		suite.T().Error(err.Error())
	}
	suite.Equal(body.Msg, `"Id" must be MD5 hash string`, fmt.Sprintf("body is %q, should be %q", w.Body.String(), id))
}

// Test CheckId middleware with invalid id characters
func (suite *CheckIdMiddlewareSuiteTester) TestInvalidIdString() {
	// GIVEN invalid id with illegal characters in id
	id := "textmsgwiththirtytwodigitslength"

	r, err := http.NewRequest("GET", "/"+id, nil)
	if err != nil {
		suite.T().Error(err.Error())
	}
	w := httptest.NewRecorder()
	// AND handle response
	suite.mux.ServeHTTP(w, r)
	// THEN response status code should be 400
	suite.Equal(w.Code, http.StatusBadRequest, fmt.Sprintf("status is %d, should be %d", w.Code, http.StatusBadRequest))
	// AND response body should contain error message
	decoder := json.NewDecoder(w.Body)
	body := struct {
		Msg string `json:"msg"`
	}{}
	err = decoder.Decode(&body)
	if err != nil {
		suite.T().Error(err.Error())
	}
	suite.Equal(body.Msg, `"Id" must be MD5 hash string`, fmt.Sprintf("body is %q, should be %q", w.Body.String(), id))
}

// TestRunCheckIdMiddlewareSuite will be run by the 'go test' command
func TestRunCheckIdMiddlewareSuite(t *testing.T) {
	Run(t, new(CheckIdMiddlewareSuiteTester))
}
