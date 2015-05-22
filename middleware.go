package main

import (
    "github.com/zenazn/goji/web"
    "gopkg.in/mgo.v2/bson"
	"net/http"
)

// Check if "Id" is bson.ObjectId
func CheckId(c *web.C, h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if fileId, ok := c.URLParams["id"]; ok {
	        if !bson.IsObjectIdHex(fileId) {
				JsonResponseMsg(w, http.StatusBadRequest, `"Id" must be bson.ObjectId of type`)
	            return
	        }
		} else {
			JsonResponseMsg(w, http.StatusBadRequest, `"Id" not set`)
            return
		}
		h.ServeHTTP(w, r)
	})
}
