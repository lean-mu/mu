package server

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api/models"
)

func (s *Server) handleAppList(c *gin.Context) {
	uri := c.Request.RequestURI
	logrus.Debugf("handleAppList %s", uri)

	ctx := c.Request.Context()

	filter := &models.AppFilter{}
	filter.Cursor, filter.PerPage = pageParams(c)
	filter.Name = c.Query("name")

	apps, err := s.datastore.GetApps(ctx, filter)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, apps)
}
