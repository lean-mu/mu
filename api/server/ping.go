package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getHandlePing(nodetype string) func(*gin.Context) {

	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"hello": "world!", "goto": "https://github.com/lean-mu/mu", "type": nodetype})
	}
}
