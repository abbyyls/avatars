package main

import (
    "gopkg.in/mgo.v2/bson"
)

type Avatar struct {
	Id          bson.ObjectId   `bson:"_id,omitempty" json:"id"`
	UrlOrigin   string          `bson:"url_origin" json:"url_origin"`
	UrlThumb    string          `bson:"url_thumb" json:"url_thumb"`
    Origin      bson.ObjectId   `bson:"origin" json:"-"`
    Thumb       bson.ObjectId   `bson:"thumb" json:"-"`
}
