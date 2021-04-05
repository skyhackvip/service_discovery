package model

import (
	"github.com/gin-gonic/gin"
	"github.com/skyhackvip/service_discovery/configs"
	"log"
)

type Nodes struct {
	nodes    []*Node
	selfAddr string
}

func NewNodes(c *configs.GlobalConfig) *Nodes {
	nodes := make([]*Node, 0, len(c.Nodes))
	for _, addr := range c.Nodes {
		n := NewNode(c, addr)
		nodes = append(nodes, n)
	}
	return &Nodes{
		nodes:    nodes,
		selfAddr: c.HttpServer,
	}
}

//同步其他节点
func (nodes *Nodes) Replicate(c *gin.Context, action configs.Action, instance *Instance) error {
	log.Println("here", len(nodes.nodes))
	if len(nodes.nodes) == 0 {
		return nil
	}
	for _, node := range nodes.nodes {
		if node.addr != nodes.selfAddr {
			//异步发送
			go nodes.action(c, node, action, instance)
		}
	}
	return nil
}

func (nodes *Nodes) action(c *gin.Context, node *Node, action configs.Action, instance *Instance) {
	switch action {
	case configs.Register:
		go node.Register(c, instance)
	case configs.Renew:
		go node.Renew(c, instance)
	case configs.Cancel:
		go node.Cancel(c, instance)
	}
}

//获取所有节点
func (nodes *Nodes) AllNodes() []*Node {
	nodeRs := make([]*Node, 0, len(nodes.nodes))
	for _, node := range nodes.nodes {
		n := &Node{
			addr:   node.addr,
			status: node.status,
		}
		nodeRs = append(nodeRs, n)
	}
	return nodeRs
}
