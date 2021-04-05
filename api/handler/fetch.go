package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/skyhackvip/service_discovery/global"
	"github.com/skyhackvip/service_discovery/model"
	"github.com/skyhackvip/service_discovery/pkg/errcode"
	"log"
	"net/http"
)

func FetchHandler(c *gin.Context) {
	log.Println("request api/fetch...")
	var req model.RequestFetch
	if e := c.ShouldBindJSON(&req); e != nil {
		err := errcode.ParamError
		c.JSON(http.StatusOK, gin.H{
			"code":    err.Code(),
			"message": err.Error(),
		})
		return
	}

	//fetch
	fetchData, err := global.Discovery.Registry.Fetch(req.Env, req.AppId, req.Status, 0)
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
		"data":    fetchData,
		"message": "",
	})
}
