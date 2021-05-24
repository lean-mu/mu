package server

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api"
)

func (s *Server) handleFnDelete(c *gin.Context) {

	uri := c.Request.RequestURI
	logrus.Debugf("handleFnDelete %s", uri)

	ctx := c.Request.Context()

	fnID := c.Param(api.FnID)

	err := s.datastore.RemoveFn(ctx, fnID)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.String(http.StatusNoContent, "")
}
