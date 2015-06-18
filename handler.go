package main

import (
	"bytes"
	"encoding/json"
	"github.com/nfnt/resize"
	"github.com/zenazn/goji/web"
	"golang.org/x/image/bmp"
	//	"gopkg.in/mgo.v2/bson"
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
	Mask []int `json:"mask"`
}

func UploadFile(c web.C, w http.ResponseWriter, r *http.Request) {
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

		// get file type
		reader, filetype, err := getFileType(file)
		if err != nil {
			JsonResponseMsg(w, http.StatusInternalServerError, `can't read the file`)
			return
		}

		// check file size
		if err = checkFileSize(file); err != nil {
			JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
			return
		}

		// check if file type is supported
		if ok := contains(supportedMediaTypes, filetype); !ok {
			JsonResponseMsg(w, http.StatusUnsupportedMediaType, `unsupported media type`)
			return
		}

		idObj := c.URLParams["id"]
		if mask.Mask != nil {
			err = InsertImageAndThumbnail(idObj, reader, filename, mask.Mask)
		} else {
			err = InsertImage(idObj, reader, filename)
		}
		if err != nil {
			JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
			return
		}

		avatar, err := GetAvatarStructById(idObj)
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

func ChangeMask(c web.C, w http.ResponseWriter, r *http.Request) {
	var avatar *Avatar
	var avatarInterface interface{}

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

	avatarInterface, err = ChangeThumbnail(c.URLParams["id"], mask.Mask)
	if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
	}
	avatar = avatarInterface.(*Avatar)

	JsonResponseFromStruct(w, http.StatusOK, avatar)
	return
}

func DeleteFile(c web.C, w http.ResponseWriter, r *http.Request) {
	err := DeleteImage(c.URLParams["id"])
	if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
	}

	JsonResponseMsg(w, http.StatusOK, `avatar was deleted successfully`)
	return
}

func GetOriginalFile(c web.C, w http.ResponseWriter, r *http.Request) {
	imageToResponse(GetOriginalImageById, c.URLParams["id"], w)
	return
}

func GetResizedFile(c web.C, w http.ResponseWriter, r *http.Request) {
	var (
		err                 error
		width, height, size uint64
		resizedImage        image.Image
	)

	if len(r.URL.Query()) == 0 {
		imageToResponse(GetThumbnailImageById, c.URLParams["id"], w)
		return
	}

	buf, err := GetThumbnailImageById(c.URLParams["id"])
	if err != nil {
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

	file := buf.(*bytes.Buffer)
	reader, filetype, err := getFileType(file)
	if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, `can't read the file`)
		return
	}

	w.Header().Set("Content-Type", filetype)

	// decode image file into image.Image
	img, _, err := image.Decode(reader)
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

func imageToResponse(fn func(string) (interface{}, error), id string, w http.ResponseWriter) {
	buf, err := fn(id)
	if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, err.Error())
		return
	}

	file := buf.(*bytes.Buffer)
	reader, filetype, err := getFileType(file)
	if err != nil {
		JsonResponseMsg(w, http.StatusInternalServerError, `can't read the file`)
		return
	}
	// set content type and other headers
	w.Header().Set("Content-Type", filetype)
	w.Header().Set("Content-Length", strconv.FormatInt(int64(reader.Len()), 10))
	io.Copy(w, reader)
	return
}
