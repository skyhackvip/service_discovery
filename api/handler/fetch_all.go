package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/skyhackvip/service_discovery/global"
	"log"
	"net/http"
)

func FetchAllHandler(c *gin.Context) {
	log.Println("request api/fetchall...")

	data := global.Discovery.Registry.FetchAll()
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "",
		"data":    data,
	})
}
