package model

import (
	"encoding/json"
)

type Response struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type FetchData struct {
	Instances       []*Instance `json:"instances"`
	LatestTimestamp int64       `json:"latest_timestamp"`
}

type ResponseFetch struct {
	Response
	Data FetchData `json:"data"`
}
