package main

import (
	"bytes"
	"errors"
	"github.com/drone/config"
	"golang.org/x/image/bmp"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
)

var (
    mgoSession     *mgo.Session
    mongoUrl = config.String("mongo-url", "mongodb://localhost/avatars")
    GridFsPrefix = config.String("mongo-gridfs-prefix", "avatars")
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

func GetOriginalAvatar(db *mgo.Database, id bson.ObjectId) (*mgo.GridFile, error) {
	return getAvatar(db, id, true)
}

func GetAvatarThumbnail(db *mgo.Database, id bson.ObjectId) (*mgo.GridFile, error) {
	return getAvatar(db, id, false)
}

func getAvatar (db *mgo.Database, id bson.ObjectId, isOrigin bool) (file *mgo.GridFile, err error) {
	var imageId bson.ObjectId
	result, err := GetAvatarStruct(db, id)
	if err != nil {
		return nil, err
	}
	if isOrigin {
		imageId = result.Origin
	} else {
		imageId = result.Thumb
	}
	file, err = db.GridFS(*GridFsPrefix).OpenId(imageId)
	if err != nil {
		return nil, err
	}

    return file, err
}

func GetAvatarStruct(db *mgo.Database, id bson.ObjectId) (result *Avatar, err error) {
	result = &Avatar{}
	err = db.C(*MongoCollection).FindId(id).One(&result)
	return result, err
}

func InsertAvatar (db *mgo.Database, file *multipart.File, filename string) (id bson.ObjectId, err error) {
	var storedFile *mgo.GridFile
	storedFile, err = db.GridFS(*GridFsPrefix).Create(filename)
    if err != nil {
        return
    }
	defer storedFile.Close()
	
	_, err = io.Copy(storedFile, *file)
    if err != nil {
        return
    }
	
	fileid := storedFile.Id().(bson.ObjectId)
	id = bson.NewObjectId()
	url := ApiUrl + "file/" + id.Hex()
	err = db.C(*MongoCollection).Insert(&Avatar{
		Id: id,
		UrlOrigin: url + "/raw",
		UrlThumb: url,
		Origin: fileid,
		Thumb: fileid,
	})

    return
}

func InsertAvatarAndThumbnail(db *mgo.Database, file *multipart.File, filename string, mask []int) (id bson.ObjectId, err error) {
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

	img, filetype, err := image.Decode(*file)
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
		return id, errors.New(`can't convert image`)
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
	id = bson.NewObjectId()
	url := ApiUrl + "file/" + id.Hex()
	
	err = db.C(*MongoCollection).Insert(&Avatar{
		Id: id,
		UrlOrigin: url + "/raw",
		UrlThumb: url,
		Origin: fileId,
		Thumb: thumbFileId,
	})
	
	return
}

func ChangeThumbnail(db *mgo.Database, id bson.ObjectId, mask []int) (err error) {
    result := Avatar{}
	if err = db.C(*MongoCollection).FindId(id).One(&result); err != nil {
		return err
	}
	if result.Origin != result.Thumb {
	    if err = db.GridFS(*GridFsPrefix).RemoveId(result.Thumb); err != nil {
	        return err
	    }
	}

	file, err := db.GridFS(*GridFsPrefix).OpenId(result.Origin)
	if err != nil {
		return err
	}

	var storedThumbFile *mgo.GridFile
	storedThumbFile, err = db.GridFS(*GridFsPrefix).Create("thumb_" + file.Name())
    if err != nil {
        return
    }
	defer storedThumbFile.Close()

	var arr []byte
	buf := bytes.NewBuffer(arr)
	_, err = io.Copy(buf, file)
	if err != nil {
        return
    }
	img, filetype, err := image.Decode(buf)
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
	return
}

func DeleteAvatar (db *mgo.Database, id bson.ObjectId) (err error) {
    result := Avatar{}
	if err = db.C(*MongoCollection).FindId(id).One(&result); err != nil {
		return err
	}
    if err = db.GridFS(*GridFsPrefix).RemoveId(result.Origin); err != nil {
        return err
    }
	if result.Origin != result.Thumb {
	    if err = db.GridFS(*GridFsPrefix).RemoveId(result.Thumb); err != nil {
	        return err
	    }
	}
    if err = db.C(*MongoCollection).RemoveId(id); err != nil {
        return err
    }
    return
}
