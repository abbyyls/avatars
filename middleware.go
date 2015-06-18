package main

import (
	"github.com/zenazn/goji/web"
	"log"
	"net/http"
	"regexp"
)

// Check if "Id" is MD5 hash string
func CheckId(c *web.C, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(c.URLParams["id"])
		if fileId, ok := c.URLParams["id"]; ok {
			re := regexp.MustCompile("[a-fA-F0-9]{32}")
			if match := re.MatchString(fileId); !match {
				JsonResponseMsg(w, http.StatusBadRequest, `"Id" must be MD5 hash string`)
				return
			}
		} else {
			JsonResponseMsg(w, http.StatusBadRequest, `"Id" not set`)
			return
		}
		h.ServeHTTP(w, r)
	})
}
