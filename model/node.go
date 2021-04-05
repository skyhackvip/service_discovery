package model

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/skyhackvip/service_discovery/configs"
	"github.com/skyhackvip/service_discovery/pkg/errcode"
	"github.com/skyhackvip/service_discovery/pkg/httputil"
	"log"
	"strconv"
)

//node is a special client
//节点注册发现、心跳保持、注销
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

func (node *Node) Register(c *gin.Context, instance *Instance) error {
	return node.call(c, node.registerURL, configs.Register, instance, nil)
}

func (node *Node) Cancel(c *gin.Context, instance *Instance) error {
	return node.call(c, node.cancelURL, configs.Cancel, instance, nil)
}

func (node *Node) Renew(c *gin.Context, instance *Instance) error {
	var res *Instance
	err := node.call(c, node.renewURL, configs.Renew, instance, &res)
	if err == errcode.ServerError {
		log.Println("node renew error") //todo
		node.status = configs.NodeStatusDown
		return err
	}
	if err == errcode.NotFound { //register
		log.Println("node renew not found, register again")
		err = node.call(c, node.registerURL, configs.Register, instance, nil)
		return err
	}
	return err
}

func (node *Node) call(c *gin.Context, uri string, action configs.Action, instance *Instance, data interface{}) error {
	fmt.Println("call other server", uri, action)
	params := make(map[string]interface{})
	params["env"] = instance.Env
	params["appid"] = instance.AppId
	params["hostname"] = instance.Hostname
	params["replication"] = "true"
	switch action {
	case configs.Register:
		params["addrs"] = instance.Addrs
		params["status"] = strconv.FormatUint(uint64(instance.Status), 10)
		params["version"] = instance.Version
		params["reg_timestamp"] = strconv.FormatInt(instance.RegTimestamp, 10)
		params["dirty_timestamp"] = strconv.FormatInt(instance.DirtyTimestamp, 10)
		params["latest_timestamp"] = strconv.FormatInt(instance.LatestTimestamp, 10)
	case configs.Renew:
		params["dirty_timestamp"] = strconv.FormatInt(instance.DirtyTimestamp, 10)
	case configs.Cancel:
		params["latest_timestamp"] = strconv.FormatInt(instance.LatestTimestamp, 10)
	}

	//request other server
	resp, err := httputil.HttpPost(uri, params)
	if err != nil {
		log.Printf("node call %v err : %v \n", uri, err)
		return err
	}
	res := Response{}
	err = json.Unmarshal([]byte(resp), &res)
	if err != nil {
		log.Printf("node call %v err : %v \n", uri, err)
		return err
	}
	if res.Code != 200 { //code
		log.Printf("uri is (%v),response ccode (%v)", uri, res.Code)
	}
	log.Println(res)
	return nil
}
