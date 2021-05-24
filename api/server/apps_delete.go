package server

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api"
)

func (s *Server) handleAppDelete(c *gin.Context) {
	ctx := c.Request.Context()

	uri := c.Request.RequestURI
	logrus.Debugf("handleAppDelete %s", uri)

	err := s.datastore.RemoveApp(ctx, c.Param(api.AppID))
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.String(http.StatusNoContent, "")
}
