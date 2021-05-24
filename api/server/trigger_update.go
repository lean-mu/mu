package server

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api"
	"github.com/lean-mu/mu/api/models"
)

func (s *Server) handleTriggerUpdate(c *gin.Context) {

	uri := c.Request.RequestURI
	logrus.Debugf("handleTriggerUpdate %s", uri)

	trigger := &models.Trigger{}

	err := c.BindJSON(trigger)
	if err != nil {
		if models.IsAPIError(err) {
			handleErrorResponse(c, err)
		} else {
			handleErrorResponse(c, models.ErrInvalidJSON)
		}
		return
	}

	pathTriggerID := c.Param(api.TriggerID)

	if trigger.ID == "" {
		trigger.ID = pathTriggerID
	} else {
		if pathTriggerID != trigger.ID {
			handleErrorResponse(c, models.ErrTriggerIDMismatch)
		}
	}

	ctx := c.Request.Context()
	triggerUpdated, err := s.datastore.UpdateTrigger(ctx, trigger)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, triggerUpdated)
}
