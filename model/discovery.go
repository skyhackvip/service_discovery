package model

import (
	"encoding/json"
	"fmt"
	"github.com/skyhackvip/service_discovery/configs"
	"github.com/skyhackvip/service_discovery/pkg/errcode"
	"github.com/skyhackvip/service_discovery/pkg/httputil"
	"log"
	"net/url"
	"sync/atomic"
	"time"
)

type Discovery struct {
	config    *configs.GlobalConfig
	protected bool
	Registry  *Registry
	Nodes     atomic.Value
}

//discovery
func NewDiscovery(config *configs.GlobalConfig) *Discovery {
	//init discovery
	dis := &Discovery{
		protected: false,
		config:    config,
		Registry:  NewRegistry(), //init registry
	}
	//new nodes
	dis.Nodes.Store(NewNodes(config))
	//instance sync
	dis.sync()
	//register current discovery
	dis.regSelf()
	//nodes sync
	go dis.nodesProtect()
	//exit protected mode
	go dis.exitProtect()
	return dis
}

//sync registry data
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

		log.Println("sync", res.Code, res.Message, len(res.Data))
		dis.protected = false
		for _, v := range res.Data {
			for _, instance := range v {
				log.Println("jieguo", instance.Addrs, instance.Hostname)
				//service registry
				dis.Registry.Register(instance, instance.LatestTimestamp)
			}
		}
	}
	nodes.SetUp()
}

//registry current as a service
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
		ticker := time.NewTicker(configs.RenewInterval) //30 second
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

//update discovery nodes list
func (dis *Discovery) nodesProtect() {
	log.Println("nodes protect......")
	var lastTimestamp int64
	for {
		log.Println("nodes protect", configs.DiscoveryAppId, lastTimestamp)
		ch, err := dis.Registry.Polls(dis.config.Env, dis.config.Hostname, []string{configs.DiscoveryAppId}, []int64{lastTimestamp})
		log.Println("nodes protect", err)
		if err != nil && err == errcode.NotModified {
			log.Println(err)
			time.Sleep(configs.NodesProtectInterval)
			continue
		}
		apps := <-ch
		log.Println("nodews protect", apps)
		fetchData, ok := apps[configs.DiscoveryAppId] //appid==Kavin.discovery all discovery node instance
		if !ok || fetchData == nil {
			return
		}
		var nodes []string
		for _, instance := range fetchData.Instances {
			for _, addr := range instance.Addrs {
				u, err := url.Parse(addr)
				if err == nil {
					nodes = append(nodes, u.Host)
				}
			}
		}
		lastTimestamp = fetchData.LatestTimestamp

		//config update new nodes
		config := new(configs.GlobalConfig)
		*config = *dis.config
		config.Nodes = nodes

		ns := NewNodes(config)
		ns.SetUp()
		dis.Nodes.Store(ns)
		log.Println("discovery changed nodes")
	}
}

//discovery exit protect after 1 minute
func (dis *Discovery) exitProtect() {
	time.Sleep(configs.ProtectTimeInterval)
	dis.protected = false
	log.Println("exit protect")
}
