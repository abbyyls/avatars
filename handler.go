package main

import (
	"encoding/json"
	"github.com/nfnt/resize"
	"github.com/zenazn/goji/web"
	"golang.org/x/image/bmp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
)

var supportedMediaTypes = []string{"image/jpeg", "image/jpg", "image/bmp", "image/png", "image/png"}

type Mask struct {
    Mask  []int  `json:"mask"`
}

func UploadFile (c web.C, w http.ResponseWriter, r *http.Request) {
	session := getSession()
	defer session.Close()
    db := session.DB("")
	
	//parse the multipart form in the request
	err := r.ParseMultipartForm(100000)
	if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
	}

	mask := Mask{}
	value := r.FormValue("config")
	if len(value) > 0 {
	    if err := json.Unmarshal([]byte(value), &mask); err != nil {
			JsonResponseMsg(w, http.StatusBadRequest, `field "config" should be json-string`)
			return
		}
		if len(mask.Mask) != 4 {
			JsonResponseMsg(w, http.StatusBadRequest, `field "config" should contain 4 integer elements`)
			return
		}
	}

	//get a ref to the parsed multipart form
	files := r.MultipartForm.File["files"]
	for i, _ := range files {
		file, err := files[i].Open()
		if err != nil {
			JsonResponseMsg(w, http.StatusInternalServerError, `can't open the file`)
			return
		}
		defer file.Close()
		
		filename := filepath.Base(files[i].Filename)
		
		// TODO: fix all checkers
		// get file type
//		buff := make([]byte, 512) // 512 bytes because http://golang.org/pkg/net/http/#DetectContentType
//        if _, err = file.Read(buff); err != nil {
//			JsonResponseMsg(w, http.StatusInternalServerError, `can't read the file`)
//			return
//        }
//        filetype := http.DetectContentType(buff)

		// check file size
//		if httperr := checkFileSize(file); httperr != nil {
//			JsonResponseMsg(w, httperr.Status(), httperr.Error())
//			return
//		}

		// check if file type is supported
//		if ok := contains(supportedMediaTypes, filetype); !ok {
//			JsonResponseMsg(w, http.StatusUnsupportedMediaType, `unsupported media type`)
//			return
//		}

		var id bson.ObjectId
		if mask.Mask != nil {
			id, err = InsertAvatarAndThumbnail(db, &file, filename, mask.Mask)
		} else {
			id, err = InsertAvatar(db, &file, filename)
		}
		if err != nil {
			JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
			return
		}

		avatar, err := GetAvatarStruct(db, id)
		if err != nil {
			JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
			return
		}
		JsonResponseFromStruct(w, http.StatusCreated, avatar)
	    return
	}

	JsonResponseMsg(w, http.StatusBadRequest, `bad request`)
    return
}

func ChangeMask (c web.C,w http.ResponseWriter,r *http.Request) {
	session := getSession()
	defer session.Close()
    db := session.DB("")
	
    decoder := json.NewDecoder(r.Body)
    var mask Mask  
    err := decoder.Decode(&mask)
	if err != nil {
		JsonResponseMsg(w, http.StatusBadRequest, `field "config" should be json-string`)
		return
	}
	if len(mask.Mask) != 4 {
		JsonResponseMsg(w, http.StatusBadRequest, `field "config" should contain 4 integer elements`)
		return
	}

	id := bson.ObjectIdHex(c.URLParams["id"])
	err = ChangeThumbnail(db, id, mask.Mask)
	if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	avatar, err := GetAvatarStruct(db, id)
	if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	JsonResponseFromStruct(w, http.StatusOK, avatar)
    return
}

func GetOriginalFile (c web.C,w http.ResponseWriter,r *http.Request) {
    session := getSession()
	defer session.Close()
    db := session.DB("")
	
	imageToResponse(db, GetOriginalAvatar, bson.ObjectIdHex(c.URLParams["id"]), w)
    return
}

func GetResizedFile (c web.C,w http.ResponseWriter,r *http.Request) {
    session := getSession()
	defer session.Close()
    db := session.DB("")
		
	var (
		file *mgo.GridFile
		err error
		width, height, size uint64
		resizedImage image.Image
	)
	
	if len(r.URL.Query()) == 0 {
		imageToResponse(db, GetAvatarThumbnail, bson.ObjectIdHex(c.URLParams["id"]), w)
		return
	}
	
	file, err = GetAvatarThumbnail(db, bson.ObjectIdHex(c.URLParams["id"]))
	defer file.Close()	
	// Only return a 404 if the error from gridfs was 'not found'.
    // If something else goes wrong, return 500.
	if err != nil {
		if err == mgo.ErrNotFound {
			JsonResponseMsg(w, http.StatusNotFound, `avatar not found`)
			return
		}
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	_, hOk := r.URL.Query()["h"]
	_, wOk := r.URL.Query()["w"]
	_, sOk := r.URL.Query()["s"]

	if hOk && wOk {
		height, err = strconv.ParseUint(r.URL.Query().Get("h"), 10, 64)
		if err != nil {
			JsonResponseMsg(w, http.StatusBadRequest, `"h" parameter should be an integer`)
			return
		}
		width, err = strconv.ParseUint(r.URL.Query().Get("w"), 10, 64)
		if err != nil {
			JsonResponseMsg(w, http.StatusBadRequest, `"w" parameter should be an integer`)
			return
		}
	} else if sOk {
		size, err = strconv.ParseUint(r.URL.Query().Get("s"), 10, 64)
		if err != nil {
			JsonResponseMsg(w, http.StatusBadRequest, `"s" parameter should be an integer`)
			return
		}
	} else {
		JsonResponseMsg(w, http.StatusBadRequest, `incorrect query parameters`)
		return
	}

	filetype := getMimeType(file)
	w.Header().Set("Content-Type", filetype)
	
    // decode image file into image.Image
    img, _, err := image.Decode(file)
    if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
    }
	
	if hOk && wOk {
		resizedImage = resize.Resize(uint(width), uint(height), img, resize.Lanczos3)
	} else if sOk {
		resizedImage = resize.Thumbnail(uint(size), uint(size), img, resize.Lanczos3)
	}
	
	// check if file type is supported
	if ok := contains(supportedMediaTypes, filetype); !ok {
		JsonResponseMsg(w, http.StatusUnsupportedMediaType, `unsupported media type`)
		return
	}
	
    switch filetype {
    case "image/jpeg", "image/jpg":
		jpeg.Encode(w, resizedImage, nil)
    case "image/bmp":
		bmp.Encode(w, resizedImage)
    case "image/png":
		png.Encode(w, resizedImage)
    case "image/gif":
		gif.Encode(w, resizedImage, nil)
    }
    return
}

func DeleteFile(c web.C, w http.ResponseWriter, r *http.Request) {
	session := getSession()
	defer session.Close()
    db := session.DB("")
	
	err := DeleteAvatar(db, bson.ObjectIdHex(c.URLParams["id"]))
	if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
	}

	JsonResponseMsg(w, http.StatusOK, `avatar was deleted successfully`)
    return
}

func imageToResponse(db *mgo.Database, fn func(*mgo.Database, bson.ObjectId) (*mgo.GridFile, error), id bson.ObjectId, w http.ResponseWriter) {
	file, err := fn(db, id)
	defer file.Close()
	// Only return a 404 if the error from gridfs was 'not found'.
    // If something else goes wrong, return 500.
	if err != nil {
		if err == mgo.ErrNotFound {
			JsonResponseMsg(w, http.StatusNotFound, `avatar not found`)
			return
		}
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	// set content type and other headers
	ctype := getMimeType(file)
	w.Header().Set("Content-Type", ctype)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size(), 10))
	w.Header().Set("ETag", file.MD5())
	io.Copy(w, file)
    return
}
