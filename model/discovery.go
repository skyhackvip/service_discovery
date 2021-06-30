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
	//new nodes from config file
	dis.Nodes.Store(NewNodes(config))

	//sync data from other nodes
	dis.initSync()

	//register discovery
	instance := dis.regSelf()
	//renew discovery
	go dis.renewTask(instance)

	//nodes perception
	go dis.nodesPerception()
	//exit protected mode
	go dis.exitProtect()
	return dis
}

//sync registry data
func (dis *Discovery) initSync() {
	nodes := dis.Nodes.Load().(*Nodes)
	for _, node := range nodes.AllNodes() {
		if node.addr == nodes.selfAddr {
			continue
		}
		uri := fmt.Sprintf("http://%s%s", node.addr, configs.FetchAllURL)
		resp, err := httputil.HttpPost(uri, nil)
		if err != nil {
			log.Println(err)
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
		if res.Code != configs.StatusOK {
			log.Printf("get from %v error : %v", uri, res.Message)
			continue
		}
		dis.protected = false
		for _, v := range res.Data {
			for _, instance := range v {
				dis.Registry.Register(instance, instance.LatestTimestamp)
			}
		}
	}
	nodes.SetUp()
}

//register current discovery node
func (dis *Discovery) regSelf() *Instance {
	log.Println("### discovery node register self when start ###")
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
	dis.Nodes.Load().(*Nodes).Replicate(configs.Register, instance) //broadcast
	return instance
}

//renew current discovery node
func (dis *Discovery) renewTask(instance *Instance) {
	now := time.Now().UnixNano()
	ticker := time.NewTicker(configs.RenewInterval) //30 second
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Println("### discovery node renew every 30s ###")
			_, err := dis.Registry.Renew(instance.Env, instance.AppId, instance.Hostname)
			if err == errcode.NotFound {
				dis.Registry.Register(instance, now)
				dis.Nodes.Load().(*Nodes).Replicate(configs.Register, instance)
			} else {
				dis.Nodes.Load().(*Nodes).Replicate(configs.Renew, instance)
			}
		}
	}
}

func (dis *Discovery) CancelSelf() {
	log.Println("### discovery node cancel self when exit ###")
	dis.Registry.Cancel(dis.config.Env, configs.DiscoveryAppId, dis.config.Hostname, time.Now().UnixNano())
	instance := &Instance{
		Env:      dis.config.Env,
		Hostname: dis.config.Hostname,
		AppId:    configs.DiscoveryAppId,
	}
	log.Println("$$$$$$$$$$$干掉自己$$$$$$$$$$$$$")
	log.Println(dis.config.Env, dis.config.Hostname, configs.DiscoveryAppId)
	dis.Nodes.Load().(*Nodes).Replicate(configs.Cancel, instance) //broadcast
}

//update discovery nodes list
func (dis *Discovery) nodesPerception() {
	var lastTimestamp int64
	ticker := time.NewTicker(configs.NodePerceptionInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			log.Println("### discovery node protect tick ###")
			log.Printf("### discovery nodes,len (%v) ###\n", len(dis.Nodes.Load().(*Nodes).AllNodes()))
			fetchData, err := dis.Registry.Fetch(dis.config.Env, configs.DiscoveryAppId, configs.NodeStatusUp, lastTimestamp)
			if err != nil || fetchData == nil {
				continue
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
			log.Printf("### discovery protect change nodes,len (%v) ###\n", len(dis.Nodes.Load().(*Nodes).AllNodes()))
		}
	}
}

//discovery exit protect after 1 minute
func (dis *Discovery) exitProtect() {
	time.Sleep(configs.ProtectTimeInterval)
	dis.protected = false
	log.Println("### discovery node exit protect after 60s ###")
}
