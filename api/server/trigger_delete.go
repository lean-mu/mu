package server

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api"
)

func (s *Server) handleTriggerDelete(c *gin.Context) {

	uri := c.Request.RequestURI
	logrus.Debugf("handleTriggerDelete %s", uri)

	ctx := c.Request.Context()

	err := s.datastore.RemoveTrigger(ctx, c.Param(api.TriggerID))
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.String(http.StatusNoContent, "")
}
