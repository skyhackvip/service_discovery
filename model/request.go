package model

//api register
type RequestRegister struct {
	Env             string   `form:"env"`
	AppId           string   `form:"appid"`
	Hostname        string   `form:"hostname"`
	Addrs           []string `form:"addrs[]"`
	Status          uint32   `form:"status"`
	Version         string   `form:"version"`
	LatestTimestamp int64    `form:"latest_timestamp"`
	DirtyTimestamp  int64    `form:"dirty_timestamp"` //other node send
	Replication     bool     `form:"replication"`     //other node send
}

//api renew heart beat
type RequestRenew struct {
	Env            string `form:"env"`
	AppId          string `form:"appid"`
	Hostname       string `form:"hostname"`
	DirtyTimestamp int64  `form:"dirty_timestamp"` //other node send
	Replication    bool   `form:"replication"`     //other node send
}

//api cancel
type RequestCancel struct {
	Env             string `form:"env"`
	AppId           string `form:"appid"`
	Hostname        string `form:"hostname"`
	LatestTimestamp int64  `form:"last_timestamp"` //other node send
	Replication     bool   `form:"replication"`    //other node send
}

//api fetch
type RequestFetch struct {
	Env    string `from:"env"`
	AppId  string `form:"appid"`
	Status uint32 `form:"status"`
}

//api fetch more
type RequestFetchs struct {
	Env    string   `form:"env"`
	AppId  []string `form:"appid"`
	Status uint32   `form:"status"`
}

//api nodes
type RequestNodes struct {
	Env string `form:"env"`
}
