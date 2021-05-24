package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api"
)

func (s *Server) handleTriggerGet(c *gin.Context) {

	uri := c.Request.RequestURI
	logrus.Debugf("handleTriggerGet %s", uri)

	ctx := c.Request.Context()

	trigger, err := s.datastore.GetTriggerByID(ctx, c.Param(api.TriggerID))

	if err != nil {
		handleErrorResponse(c, err)
		return
	}
	app, err := s.datastore.GetAppByID(ctx, trigger.AppID)

	if err != nil {
		handleErrorResponse(c, fmt.Errorf("unexpected error - trigger app not available: %s", err))
		return
	}

	trigger, err = s.triggerAnnotator.AnnotateTrigger(c, app, trigger)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, trigger)
}
