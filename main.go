package main

import (
	"github.com/drone/config"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"net/http"
)

const (
	BaseUrl    = `/`
	ApiUrl     = `api/v1/`
	BaseApiUrl = BaseUrl + ApiUrl
)

var (
	Listen = config.String("http", "0.0.0.0:4567")
)

// Options automatically return an appropriate "Allow" header when the
// request method is OPTIONS and the request would have otherwise been 404'd.
func Options(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, api_key, Authorization")
			w.Header().Set("Allow", "OPTIONS,HEAD,GET,POST,PUT,DELETE,OPTIONS")
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func main() {
	config.SetPrefix("AV_")
	config.Parse("")

	mux := web.New()
	mux.Use(middleware.Logger) // TODO: remove??
	mux.Use(Options)

	// NOTE: "RouterWithId" router needs for using URL parameters in CheckId middleware.
	// Goji can't bind URL parameters until after the middleware stack runs.
	// https://github.com/zenazn/goji/issues/32#issuecomment-46124240
	RouterWithId := web.New()
	RouterWithId.Use(CheckId)
	RouterWithId.Post(BaseApiUrl+"file/:id", UploadFile)
	RouterWithId.Put(BaseApiUrl+"file/:id", ChangeMask)
	RouterWithId.Delete(BaseApiUrl+"file/:id", DeleteFile)
	RouterWithId.Get(BaseApiUrl+"file/:id", GetResizedFile)
	RouterWithId.Get(BaseApiUrl+"file/:id/raw", GetOriginalFile)

	mux.Handle(BaseApiUrl+"file/:id", RouterWithId)
	mux.Handle(BaseApiUrl+"file/:id/*", RouterWithId)

	http.Handle(BaseApiUrl, mux)

	http.Handle(BaseUrl, http.FileServer(http.Dir("app")))

	panic(http.ListenAndServe(*Listen, nil))
}
