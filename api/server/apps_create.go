package server

import (
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api/models"
)

func (s *Server) handleAppCreate(c *gin.Context) {

	uri := c.Request.RequestURI
	logrus.Debugf("handleAppCreate %s", uri)

	ctx := c.Request.Context()

	app := &models.App{}

	err := c.BindJSON(app)
	if err != nil {
		if models.IsAPIError(err) {
			handleErrorResponse(c, err)
		} else {
			handleErrorResponse(c, models.ErrInvalidJSON)
		}
		return
	}

	app, err = s.datastore.InsertApp(ctx, app)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	c.JSON(http.StatusOK, app)
}
