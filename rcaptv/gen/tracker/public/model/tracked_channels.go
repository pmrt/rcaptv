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

type TrackedChannels struct {
	BcID               string `sql:"primary_key"`
	BcDisplayName      string
	BcUsername         string
	BcType             Broadcastertype
	PpURL              *string
	OfflinePpURL       *string
	TrackedSince       *time.Time
	SeenInactiveCount  *int32
	EnabledStatus      *bool
	LastModifiedStatus *time.Time
	PriorityLvl        *int16
}
