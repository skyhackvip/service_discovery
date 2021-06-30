package model

import (
	//"github.com/gin-gonic/gin"
	"github.com/skyhackvip/service_discovery/configs"
	"log"
)

type Nodes struct {
	nodes    []*Node
	selfAddr string
}

//new nodes
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

//replicate to other nodes
func (nodes *Nodes) Replicate(action configs.Action, instance *Instance) error {
	log.Printf("$$$$$$$$$$ %v $$$$$$$$$$$$\n", len(nodes.nodes))
	if len(nodes.nodes) == 0 {
		return nil
	}
	for _, node := range nodes.nodes {
		if node.addr != nodes.selfAddr {
			log.Printf("### replicate node(%v),action(%v),hostname(%v) ###\n", node.addr, action, instance.Hostname)
			go nodes.action(node, action, instance)
		}
	}
	return nil
}

//use node action
func (nodes *Nodes) action(node *Node, action configs.Action, instance *Instance) {
	switch action {
	case configs.Register:
		go node.Register(instance)
	case configs.Renew:
		go node.Renew(instance)
	case configs.Cancel:
		go node.Cancel(instance)
	}
}

//get all nodes
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

//set up current node
func (nodes *Nodes) SetUp() {
	for _, node := range nodes.nodes {
		if node.addr == nodes.selfAddr {
			node.status = configs.NodeStatusUp
		}
	}
}
