package main

import (
	"bytes"
	"io"
	"os"
	"testing"
)

type MongoSuiteTester struct {
	BaseSuiteWithDB

	id        string // user id
	filename  string // original file name
	thumbname string // thumbnail file name
	image     []byte
	thumb     []byte
}

// Settings for suite
func (suite *MongoSuiteTester) SetupSuite() {
	// INIT set file names
	suite.filename = "test_picture.png"
	suite.thumbname = "test_thumbnail.png"
	// AND store raw image as byte array
	buf := bytes.NewBuffer(nil)
	f, err := os.Open(suite.filename)
	if err != nil {
		suite.T().Error(err.Error())
	}
	io.Copy(buf, f)
	f.Close()
	suite.image = buf.Bytes()
	// AND store thumbnail image as byte array
	buf = bytes.NewBuffer(nil)
	f, err = os.Open(suite.thumbname)
	if err != nil {
		suite.T().Error(err.Error())
	}
	io.Copy(buf, f)
	f.Close()
	suite.thumb = buf.Bytes()
}

// Settings for each test
func (suite *MongoSuiteTester) SetupTest() {
	// INIT set random user id
	suite.id = RandomMD5()
}

// Test inserting image
func (suite *MongoSuiteTester) TestInsertImage() {
	// GIVEN 'file upload first time' flag
	isNew := true

	// WHEN I upload the file
	err := InsertImage(suite.id, suite.image, suite.filename, isNew)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND get Avatar struct
	avatar, err := GetAvatarStructById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// THEN user id should equal avatar id
	suite.Equal(suite.id, avatar.Id)
	// AND original image id should equal thumbnail image id
	suite.Equal(avatar.Origin, avatar.Thumb)

	// WHEN I get original image from database
	buf, err := GetOriginalImageById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	file := buf.(*bytes.Buffer)
	// THEN original image should equal stored file
	suite.Equal(file.Bytes(), suite.image)

	// WHEN I get thumbnail image from database
	buf, err = GetThumbnailImageById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	file = buf.(*bytes.Buffer)
	// THEN thumbnail image should equal stored file
	suite.Equal(file.Bytes(), suite.image)
}

// Test inserting image and replacing with new image
func (suite *MongoSuiteTester) TestInsertImageAndReplaceWithNewImage() {
	// GIVEN 'file upload first time' flag
	isNew := true

	// WHEN I upload the file
	err := InsertImage(suite.id, suite.image, suite.filename, isNew)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND upload thumbnail file to replace stored image
	isNew = false
	err = InsertImage(suite.id, suite.thumb, suite.thumbname, isNew)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND get Avatar struct
	avatar, err := GetAvatarStructById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// THEN user id should equal avatar id
	suite.Equal(suite.id, avatar.Id)
	// AND original image id should equal thumbnail image id
	suite.Equal(avatar.Origin, avatar.Thumb)

	// WHEN I get original image from database
	buf, err := GetOriginalImageById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	file := buf.(*bytes.Buffer)
	// THEN original image should equal stored thumbnail file
	suite.Equal(file.Bytes(), suite.thumb)

	// WHEN I get thumbnail image from database
	buf, err = GetThumbnailImageById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	file = buf.(*bytes.Buffer)
	// THEN thumbnail image should equal stored thumbnail file
	suite.Equal(file.Bytes(), suite.thumb)
}

// Test inserting image and thumbnail
func (suite *MongoSuiteTester) TestInsertImageAndThumbnail() {
	// GIVEN 'file upload first time' flag
	isNew := true
	// AND mask struct
	mask := Mask{Mask: []int{70, 15, 250, 130}}

	// WHEN I upload the file with given mask
	err := InsertImageAndThumbnail(suite.id, suite.image, suite.filename, mask.Mask, isNew)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND get Avatar struct
	avatar, err := GetAvatarStructById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// THEN user id should equal avatar id
	suite.Equal(suite.id, avatar.Id)
	// AND original image id should not equal thumbnail image id
	suite.NotEqual(avatar.Origin, avatar.Thumb)

	// WHEN I get original image from database
	buf, err := GetOriginalImageById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	file := buf.(*bytes.Buffer)
	// THEN original image should equal stored file
	suite.Equal(file.Bytes(), suite.image)

	// WHEN I get thumbnail image from database
	buf, err = GetThumbnailImageById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	file = buf.(*bytes.Buffer)
	// THEN thumbnail image should equal thumbnail test file
	suite.Equal(file.Bytes(), suite.thumb)
}

// Test inserting image and thumbnail and replacing with new image
func (suite *MongoSuiteTester) TestInsertImageAndThumbnailAndReplaceWithNewImage() {
	// GIVEN 'file upload first time' flag
	isNew := true
	// AND mask struct
	mask := Mask{Mask: []int{70, 15, 250, 130}}

	// WHEN I upload the file with given mask
	err := InsertImageAndThumbnail(suite.id, suite.image, suite.filename, mask.Mask, isNew)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND upload thumbnail file to replace stored image with given mask
	isNew = false
	err = InsertImageAndThumbnail(suite.id, suite.thumb, suite.thumbname, mask.Mask, isNew)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND get Avatar struct
	avatar, err := GetAvatarStructById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// THEN user id should equal avatar id
	suite.Equal(suite.id, avatar.Id)
	// AND original image id should not equal thumbnail image id
	suite.NotEqual(avatar.Origin, avatar.Thumb)

	// WHEN I get original image from database
	buf, err := GetOriginalImageById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	file := buf.(*bytes.Buffer)
	// THEN original image should equal stored thumbnail file
	suite.Equal(file.Bytes(), suite.thumb)

	// WHEN I get thumbnail image from database
	buf, err = GetThumbnailImageById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	file = buf.(*bytes.Buffer)
	// THEN thumbnail image should not equal thumbnail test file
	suite.NotEqual(file.Bytes(), suite.thumb)
}

// Test changing image mask
func (suite *MongoSuiteTester) TestChangeMask() {
	// GIVEN 'file upload first time' flag
	isNew := true
	// AND mask struct
	mask := Mask{Mask: []int{70, 15, 250, 130}}

	// WHEN I upload the file with given mask
	err := InsertImageAndThumbnail(suite.id, suite.image, suite.filename, mask.Mask, isNew)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND get Avatar struct
	avatar, err := GetAvatarStructById(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND change mask to stored file
	mask = Mask{Mask: []int{10, 10, 20, 20}}
	avatarInterface, err := ChangeThumbnail(suite.id, mask.Mask)
	avatarNew := avatarInterface.(*Avatar)
	// THEN user id before changing should equal user id after changing
	suite.Equal(avatar.Id, avatarNew.Id)
	// THEN original file id before changing should equal original file id after changing
	suite.Equal(avatar.Origin, avatarNew.Origin)
	// THEN thumbnail id before changing should not equal thumbnail id after changing
	suite.NotEqual(avatar.Thumb, avatarNew.Thumb)
}

// Test deleting image
func (suite *MongoSuiteTester) TestDeleteImage() {
	// GIVEN 'file upload first time' flag
	isNew := true

	// WHEN I upload the file
	err := InsertImage(suite.id, suite.image, suite.filename, isNew)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND delete this file
	err = DeleteImage(suite.id)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND get Avatar struct
	_, err = GetAvatarStructById(suite.id)
	// THEN 'not found' error should be raised
	suite.Equal(err.Error(), "not found")
}

// Test errors raising
func (suite *MongoSuiteTester) TestErrors() {
	// WHEN I trying to delete file which is not existed
	err := DeleteImage(suite.id)
	// THEN 'not found' error should be raised
	if err.Error() != "not found" {
		suite.T().Error(err.Error())
	}

	// WHEN I upload the file
	isNew := true
	err = InsertImage(suite.id, suite.image, suite.filename, isNew)
	if err != nil {
		suite.T().Error(err.Error())
	}
	// AND trying to upload another file with 'new' flag
	isNew = true
	err = InsertImage(suite.id, suite.thumb, suite.thumbname, isNew)
	// THEN 'already exists' error should be raised
	if err.Error() != "avatar for this user is already exists" {
		suite.T().Error(err.Error())
	}
}

// TestRunMongoSuite will be run by the 'go test' command
func TestRunMongoSuite(t *testing.T) {
	Run(t, new(MongoSuiteTester))
}
