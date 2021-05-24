package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api/common"
	"github.com/lean-mu/mu/api/models"
)

func (s *Server) handleTriggerCreate(c *gin.Context) {

	uri := c.Request.RequestURI
	logrus.Debugf("handleTriggerCreate %s", uri)

	ctx := c.Request.Context()
	trigger := &models.Trigger{}
	log := common.Logger(ctx)

	err := c.BindJSON(trigger)
	if err != nil {
		if models.IsAPIError(err) {
			handleErrorResponse(c, err)
		} else {
			handleErrorResponse(c, models.ErrInvalidJSON)
		}
		return
	}

	triggerCreated, err := s.datastore.InsertTrigger(ctx, trigger)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	app, err := s.datastore.GetAppByID(ctx, triggerCreated.AppID)
	if err != nil {
		log.Debugln(fmt.Errorf("unexpected error - trigger app not available: %s", err))
		c.JSON(http.StatusOK, triggerCreated)
		return
	}

	triggerAnnotated, err := s.triggerAnnotator.AnnotateTrigger(c, app, triggerCreated)
	if err != nil {
		log.Debugln("Failed to annotate trigger on cration")
		c.JSON(http.StatusOK, triggerCreated)
		return
	}

	c.JSON(http.StatusOK, triggerAnnotated)
}
