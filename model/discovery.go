package model

import (
	"encoding/json"
	"fmt"
	"github.com/skyhackvip/service_discovery/configs"
	"github.com/skyhackvip/service_discovery/pkg/httputil"
	"log"
	"sync/atomic"
	"time"
)

type Discovery struct {
	config    *configs.GlobalConfig
	protected bool
	Registry  *Registry
	Nodes     atomic.Value
}

func NewDiscovery(config *configs.GlobalConfig) *Discovery {
	//init discovery
	dis := &Discovery{
		protected: false,
		config:    config,
		Registry:  NewRegistry(), //init registry
	}
	//other nodes
	dis.Nodes.Store(NewNodes(config))
	//sync
	dis.sync()
	//register current discovery
	dis.regSelf()
	//exit protected mode
	go dis.exitProtect()
	return dis
}

//同步注册表
func (dis *Discovery) sync() {
	fmt.Println("sync")
	nodes := dis.Nodes.Load().(*Nodes)
	for _, node := range nodes.AllNodes() {
		if node.addr == nodes.selfAddr {
			continue
		}
		uri := fmt.Sprintf("http://%s%s", node.addr, configs.FetchAllURL)

		resp, err := httputil.HttpPost(uri, nil)
		if err != nil {
			fmt.Println(err)
			continue
		}
		var res struct {
			Code    int                    `json:"code"`
			Message string                 `json:"message"`
			Data    map[string][]*Instance `json:"data"`
		}
		err = json.Unmarshal([]byte(resp), &res)
		if err != nil {
			log.Printf("get from %v error : %v", uri, err)
			continue
		}
		if res.Code != 200 { //todo success
			log.Printf("get from %v error : %v", uri, res.Message)
			continue
		}

		fmt.Println("获取到结果", res.Code, res.Message, len(res.Data))
		dis.protected = false
		for _, v := range res.Data {
			for _, instance := range v {
				fmt.Println("jieguo", instance.Addrs, instance.Hostname)
				//service registry
				dis.Registry.Register(instance, instance.LatestTimestamp)
			}
		}

	}
}

//注册自身，并30s发一次心跳
func (dis *Discovery) regSelf() {
	//register
	now := time.Now().UnixNano()
	instance := &Instance{
		Env:             dis.config.Env,
		Hostname:        dis.config.Hostname,
		AppId:           configs.DiscoveryAppId,
		Addrs:           []string{"http://" + dis.config.HttpServer},
		Status:          configs.NodeStatusUp,
		RegTimestamp:    now,
		UpTimestamp:     now,
		LatestTimestamp: now,
		RenewTimestamp:  now,
		DirtyTimestamp:  now,
	}
	dis.Registry.Register(instance, now)

	//start ticker to keep alive
	go func() {
		ticker := time.NewTicker(30 * time.Second) //30 todo
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Println("renew")
				//renew
				if _, err := dis.Registry.Renew(instance.Env, instance.AppId, instance.Hostname); err != nil { //err == ecode.NothingFound
					//if renew error, register
					dis.Registry.Register(instance, now) //now 未更新最新？？
				}
			}
		}
	}()
}

//离开保护模式
func (dis *Discovery) exitProtect() {
	time.Sleep(time.Second * 60)
	dis.protected = false
}
