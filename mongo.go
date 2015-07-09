package main

import (
	"bytes"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/drone/config"
	"golang.org/x/image/bmp"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	mgoSession      *mgo.Session
	mongoUrl        = config.String("mongo-url", "mongodb://localhost/avatars")
	GridFsPrefix    = config.String("mongo-gridfs-prefix", "avatars")
	MongoDatabase   = config.String("mongo-database-name", "")
	MongoCollection = config.String("mongo-collection-name", "avatars")
)

// Connect to MongoDB and returns session clone.
func getSession() *mgo.Session {
	if mgoSession == nil {
		var err error
		mgoSession, err = mgo.Dial(*mongoUrl)
		if err != nil {
			panic(err)
		}
	}
	return mgoSession.Clone()
}

func getObjectWithDatabase(fn func(*mgo.Database) (interface{}, error)) (interface{}, error) {
	session := getSession()
	defer session.Close()
	db := session.DB(*MongoDatabase)
	return fn(db)
}

func withDatabase(fn func(*mgo.Database) error) error {
	session := getSession()
	defer session.Close()
	db := session.DB(*MongoDatabase)
	return fn(db)
}

func withCollection(collection string, fn func(*mgo.Collection) error) error {
	session := getSession()
	defer session.Close()
	c := session.DB(*MongoDatabase).C(collection)
	return fn(c)
}

func GetAvatarStructById(id string) (searchResult *Avatar, err error) {
	searchResult, err = getAvatarStruct(*MongoCollection, bson.M{"_id": id})
	return
}

func getAvatarStruct(collectionName string, q interface{}) (searchResult *Avatar, err error) {
	searchResult = &Avatar{}
	query := func(c *mgo.Collection) error {
		err := c.Find(q).One(&searchResult)
		return err
	}
	search := func() error {
		return withCollection(collectionName, query)
	}
	err = search()
	if err != nil {
		return
	}
	return
}

func GetOriginalImageById(id string) (file interface{}, err error) {
	return getImageById(id, true)
}

func GetThumbnailImageById(id string) (file interface{}, err error) {
	return getImageById(id, false)
}

func getImageById(id string, isOrigin bool) (file interface{}, err error) {
	query := func(db *mgo.Database) (interface{}, error) {
		result := &Avatar{}
		err = db.C(*MongoCollection).FindId(id).One(&result)
		if err != nil {
			return nil, err
		}

		var imageId bson.ObjectId
		if isOrigin {
			imageId = result.Origin
		} else {
			imageId = result.Thumb
		}

		gridFile, err := db.GridFS(*GridFsPrefix).OpenId(imageId)
		if err != nil {
			return nil, err
		}

		var arr []byte
		buf := bytes.NewBuffer(arr)
		_, err = io.Copy(buf, gridFile)

		return buf, err
	}
	search := func() (interface{}, error) {
		return getObjectWithDatabase(query)
	}
	file, err = search()
	if err != nil {
		return
	}
	return
}

func InsertImage(id string, file *bytes.Reader, filename string) (err error) {
	query := func(db *mgo.Database) (err error) {
		var storedFile *mgo.GridFile
		storedFile, err = db.GridFS(*GridFsPrefix).Create(filename)
		if err != nil {
			return
		}
		defer storedFile.Close()

		_, err = io.Copy(storedFile, file)
		if err != nil {
			return
		}

		fileid := storedFile.Id().(bson.ObjectId)
		url := ApiUrl + "file/" + id
		err = db.C(*MongoCollection).Insert(&Avatar{
			Id:        id,
			UrlOrigin: url + "/raw",
			UrlThumb:  url,
			Origin:    fileid,
			Thumb:     fileid,
		})
		return err
	}
	search := func() (err error) {
		return withDatabase(query)
	}
	err = search()
	if err != nil {
		return
	}
	return
}

func InsertImageAndThumbnail(id string, file *bytes.Reader, filename string, mask []int) (err error) {
	query := func(db *mgo.Database) (err error) {
		var storedFile, storedThumbFile *mgo.GridFile
		storedFile, err = db.GridFS(*GridFsPrefix).Create(filename)
		if err != nil {
			return
		}
		defer storedFile.Close()

		storedThumbFile, err = db.GridFS(*GridFsPrefix).Create("thumb_" + filename)
		if err != nil {
			return
		}
		defer storedThumbFile.Close()

		img, filetype, err := image.Decode(file)
		if err != nil {
			return
		}

		rect := image.Rect(mask[0], mask[1], mask[2], mask[3])

		var thumb image.Image
		switch pic := img.(type) {
		case *image.NRGBA:
			thumb = pic.SubImage(rect)
		case *image.NRGBA64:
			thumb = pic.SubImage(rect)
		case *image.RGBA:
			thumb = pic.SubImage(rect)
		case *image.RGBA64:
			thumb = pic.SubImage(rect)
		case *image.Gray:
			thumb = pic.SubImage(rect)
		case *image.Gray16:
			thumb = pic.SubImage(rect)
		case *image.YCbCr:
			thumb = pic.SubImage(rect)
		case *image.Paletted:
			thumb = pic.SubImage(rect)
		default:
			return errors.New(`can't convert image`)
		}

		switch filetype {
		case "jpeg", "jpg":
			jpeg.Encode(storedThumbFile, thumb, nil)
			jpeg.Encode(storedFile, img, nil)
		case "bmp":
			bmp.Encode(storedThumbFile, thumb)
			bmp.Encode(storedFile, img)
		case "png":
			png.Encode(storedThumbFile, thumb)
			png.Encode(storedFile, img)
		case "gif":
			gif.Encode(storedThumbFile, thumb, nil)
			gif.Encode(storedFile, img, nil)
		}

		fileId := storedFile.Id().(bson.ObjectId)
		thumbFileId := storedThumbFile.Id().(bson.ObjectId)
		url := ApiUrl + "file/" + id

		err = db.C(*MongoCollection).Insert(&Avatar{
			Id:        id,
			UrlOrigin: url + "/raw",
			UrlThumb:  url,
			Origin:    fileId,
			Thumb:     thumbFileId,
		})
		return err
	}
	search := func() (err error) {
		return withDatabase(query)
	}
	err = search()
	if err != nil {
		return
	}
	return
}

func ChangeThumbnail(id string, mask []int) (result interface{}, err error) {
	query := func(db *mgo.Database) (interface{}, error) {
		var err error
		searchResult := &Avatar{}
		if err = db.C(*MongoCollection).FindId(id).One(&searchResult); err != nil {
			return nil, err
		}
		if searchResult.Origin != searchResult.Thumb {
			if err = db.GridFS(*GridFsPrefix).RemoveId(searchResult.Thumb); err != nil {
				return nil, err
			}
		}

		file, err := db.GridFS(*GridFsPrefix).OpenId(searchResult.Origin)
		if err != nil {
			return nil, err
		}

		var storedThumbFile *mgo.GridFile
		storedThumbFile, err = db.GridFS(*GridFsPrefix).Create("thumb_" + file.Name())
		if err != nil {
			return nil, err
		}
		defer storedThumbFile.Close()

		var arr []byte
		buf := bytes.NewBuffer(arr)
		_, err = io.Copy(buf, file)
		if err != nil {
			return nil, err
		}
		img, filetype, err := image.Decode(buf)
		if err != nil {
			return nil, err
		}
		rect := image.Rect(mask[0], mask[1], mask[2], mask[3])

		var thumb image.Image

		switch pic := img.(type) {
		case *image.NRGBA:
			thumb = pic.SubImage(rect)
		case *image.NRGBA64:
			thumb = pic.SubImage(rect)
		case *image.RGBA:
			thumb = pic.SubImage(rect)
		case *image.RGBA64:
			thumb = pic.SubImage(rect)
		case *image.Gray:
			thumb = pic.SubImage(rect)
		case *image.Gray16:
			thumb = pic.SubImage(rect)
		case *image.YCbCr:
			thumb = pic.SubImage(rect)
		case *image.Paletted:
			thumb = pic.SubImage(rect)
		default:
			return nil, errors.New(`can't convert image`)
		}

		switch filetype {
		case "jpeg", "jpg":
			jpeg.Encode(storedThumbFile, thumb, nil)
		case "bmp":
			bmp.Encode(storedThumbFile, thumb)
		case "png":
			png.Encode(storedThumbFile, thumb)
		case "gif":
			gif.Encode(storedThumbFile, thumb, nil)
		}

		thumbFileId := storedThumbFile.Id().(bson.ObjectId)
		change := bson.M{"$set": bson.M{"thumb": thumbFileId}}
		err = db.C(*MongoCollection).UpdateId(id, change)
		if err != nil {
			return nil, err
		}

		result := &Avatar{}
		err = db.C(*MongoCollection).FindId(id).One(&result)

		return result, err
	}
	search := func() (result interface{}, err error) {
		return getObjectWithDatabase(query)
	}
	result, err = search()
	if err != nil {
		return
	}
	return
}

func DeleteImage(id string) (err error) {
	query := func(db *mgo.Database) (err error) {
		result := Avatar{}
		if err = db.C(*MongoCollection).FindId(id).One(&result); err != nil {
			return
		}
		if err = db.GridFS(*GridFsPrefix).RemoveId(result.Origin); err != nil {
			return
		}
		if result.Origin != result.Thumb {
			if err = db.GridFS(*GridFsPrefix).RemoveId(result.Thumb); err != nil {
				return
			}
		}
		if err = db.C(*MongoCollection).RemoveId(id); err != nil {
			return
		}
		return
	}
	search := func() (err error) {
		return withDatabase(query)
	}
	err = search()
	if err != nil {
		return
	}
	return
}
