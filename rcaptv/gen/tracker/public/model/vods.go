//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package model

import (
	"time"
)

type Vods struct {
	VideoID         string `sql:"primary_key"`
	StreamID        string
	BcID            string
	CreatedAt       time.Time
	PublishedAt     time.Time
	DurationSeconds int32
	Lang            string
	ThumbnailURL    string
	Title           string
	ViewCount       int32
}
