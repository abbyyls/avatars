package main

import (
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/zenazn/goji/web"
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

func SetHeaders(c *web.C, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Allow", "HEAD,GET,POST,PUT,DELETE,OPTIONS,PATCH")
		w.Header().Add("X-Frame-Options", "DENY")
		w.Header().Add("X-Content-Type-Options", "nosniff")
		w.Header().Add("X-XSS-Protection", "1; mode=block")
		w.Header().Add("Cache-Control", "no-cache")
		w.Header().Add("Cache-Control", "no-store")
		w.Header().Add("Cache-Control", "max-age=0")
		w.Header().Add("Cache-Control", "must-revalidate")
		w.Header().Add("Cache-Control", "value")
		w.Header().Add("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
		h.ServeHTTP(w, r)
	})
}

// Options automatically return an appropriate "Allow" header when the
// request method is OPTIONS and the request would have otherwise been 404'd.
func Options(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS,PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, api_key, Authorization")
			w.Header().Set("Allow", "HEAD,GET,POST,PUT,DELETE,OPTIONS,PATCH")
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
