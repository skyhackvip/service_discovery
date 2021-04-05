package model

import (
	"time"
)

type Instance struct {
	Env      string   `json:"env"`
	AppId    string   `json:"appid"`
	Hostname string   `json:"hostname"`
	Addrs    []string `json:"addrs"`
	Version  string   `json:"version"`
	Status   uint32   `json:"status"`

	RegTimestamp    int64 `json:"reg_timestamp"`
	UpTimestamp     int64 `json:"up_timestamp"`
	RenewTimestamp  int64 `json:"renew_timestamp"`
	DirtyTimestamp  int64 `json:"dirty_timestamp"`
	LatestTimestamp int64 `json:"latest_timestamp"`
}

func NewInstance(req *RequestRegister) *Instance {
	now := time.Now().UnixNano()
	instance := &Instance{
		Env:             req.Env,
		AppId:           req.AppId,
		Hostname:        req.Hostname,
		Addrs:           req.Addrs,
		Version:         req.Version,
		Status:          req.Status,
		RegTimestamp:    now,
		UpTimestamp:     now,
		RenewTimestamp:  now,
		DirtyTimestamp:  now,
		LatestTimestamp: now,
	}
	return instance
}
