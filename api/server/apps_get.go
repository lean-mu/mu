package server

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api"
)

func (s *Server) handleAppGet(c *gin.Context) {

	uri := c.Request.RequestURI
	logrus.Debugf("handleAppGet %s", uri)

	ctx := c.Request.Context()

	appId := c.Param(api.AppID)
	app, err := s.datastore.GetAppByID(ctx, appId)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, app)
}
