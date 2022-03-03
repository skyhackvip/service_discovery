package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/skyhackvip/service_discovery/configs"
	"github.com/skyhackvip/service_discovery/global"
	"github.com/skyhackvip/service_discovery/model"
	"github.com/skyhackvip/service_discovery/pkg/errcode"
	"log"
	"net/http"
)

func NodesHandler(c *gin.Context) {
	log.Println("request api/nodes...")
	var req model.RequestNodes
	if e := c.ShouldBindJSON(&req); e != nil {
		err := errcode.ParamError
		c.JSON(http.StatusOK, gin.H{
			"code":    err.Code(),
			"message": err.Error(),
		})
		return
	}

	fetchData, err := global.Discovery.Registry.Fetch(req.Env, configs.DiscoveryAppId, configs.NodeStatusUp, 0)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    err.Code(),
			"data":    "",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    fetchData,
	})
}
