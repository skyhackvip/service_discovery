package model

import (
	"encoding/json"
	"fmt"
	//	"github.com/gin-gonic/gin"
	"github.com/skyhackvip/service_discovery/configs"
	"github.com/skyhackvip/service_discovery/pkg/errcode"
	"github.com/skyhackvip/service_discovery/pkg/httputil"
	"log"
	"strconv"
	"time"
)

//node is a special client
type Node struct {
	config      *configs.Config
	addr        string
	status      int
	registerURL string
	cancelURL   string
	renewURL    string
}

func NewNode(config *configs.GlobalConfig, addr string) *Node {
	return &Node{
		addr:        addr,
		status:      configs.NodeStatusDown, //default set down
		registerURL: fmt.Sprintf("http://%s%s", addr, configs.RegisterURL),
		cancelURL:   fmt.Sprintf("http://%s%s", addr, configs.CancelURL),
		renewURL:    fmt.Sprintf("http://%s%s", addr, configs.RenewURL),
	}
}

func (node *Node) Register(instance *Instance) error {
	return node.call(node.registerURL, configs.Register, instance, nil)
}

func (node *Node) Cancel(instance *Instance) error {
	return node.call(node.cancelURL, configs.Cancel, instance, nil)
}

func (node *Node) Renew(instance *Instance) error {
	var res *Instance
	err := node.call(node.renewURL, configs.Renew, instance, &res)
	if err == errcode.ServerError {
		log.Printf("node call %s ! renew error %s \n", node.renewURL, err)
		node.status = configs.NodeStatusDown //node down
		return err
	}
	if err == errcode.NotFound { //register
		log.Printf("node call %s ! renew not found, register again \n", node.renewURL)
		return node.call(node.registerURL, configs.Register, instance, nil)
	}
	if err == errcode.Conflict && res != nil {
		return node.call(node.registerURL, configs.Register, res, nil)
	}
	return err
}

func (node *Node) call(uri string, action configs.Action, instance *Instance, data interface{}) error {
	params := make(map[string]interface{})
	params["env"] = instance.Env
	params["appid"] = instance.AppId
	params["hostname"] = instance.Hostname
	params["replication"] = true //broadcast stop here
	switch action {
	case configs.Register:
		params["addrs"] = instance.Addrs
		params["status"] = instance.Status
		params["version"] = instance.Version
		params["reg_timestamp"] = strconv.FormatInt(instance.RegTimestamp, 10)
		params["dirty_timestamp"] = strconv.FormatInt(instance.DirtyTimestamp, 10)
		params["latest_timestamp"] = strconv.FormatInt(instance.LatestTimestamp, 10)
	case configs.Renew:
		params["dirty_timestamp"] = strconv.FormatInt(instance.DirtyTimestamp, 10)
		params["renew_timestamp"] = time.Now().UnixNano()
	case configs.Cancel:
		params["latest_timestamp"] = strconv.FormatInt(instance.LatestTimestamp, 10)
	}
	//request other server
	resp, err := httputil.HttpPost(uri, params)
	if err != nil {
		log.Println(err)
		return err
	}
	res := Response{}
	err = json.Unmarshal([]byte(resp), &res)
	if err != nil {
		log.Println(err)
		return err
	}
	if res.Code != configs.StatusOK { //code!=200
		log.Printf("uri is (%v),response code (%v)", uri, res.Code)
		json.Unmarshal([]byte(res.Data), data)
		return errcode.Conflict
	}
	return nil
}
