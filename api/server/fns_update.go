package server

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api"
	"github.com/lean-mu/mu/api/models"
)

func (s *Server) handleFnUpdate(c *gin.Context) {

	uri := c.Request.RequestURI
	logrus.Debugf("handleFnUpdate %s", uri)

	ctx := c.Request.Context()

	fn := &models.Fn{}
	err := c.BindJSON(fn)
	if err != nil {
		if !models.IsAPIError(err) {
			err = models.ErrInvalidJSON
		}
		handleErrorResponse(c, err)
		return
	}

	pathFnID := c.Param(api.FnID)

	if fn.ID == "" {
		fn.ID = pathFnID
	} else {
		if pathFnID != fn.ID {
			handleErrorResponse(c, models.ErrFnsIDMismatch)
		}
	}

	fnUpdated, err := s.datastore.UpdateFn(ctx, fn)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, fnUpdated)
}
