package server

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/lean-mu/mu/api/models"
)

func (s *Server) handleFnList(c *gin.Context) {

	uri := c.Request.RequestURI
	logrus.Debugf("handleFnList %s", uri)

	ctx := c.Request.Context()

	var filter models.FnFilter
	filter.Cursor, filter.PerPage = pageParams(c)
	filter.AppID = c.Query("app_id")
	filter.Name = c.Query("name")

	fns, err := s.datastore.GetFns(ctx, &filter)
	if err != nil {
		handleErrorResponse(c, err)
		return
	}

	// Annotate the outbound fns

	// this is fairly cludgy bit hard to do in datastore middleware confidently
	appCache := make(map[string]*models.App)

	for idx, f := range fns.Items {
		app, ok := appCache[f.AppID]
		if !ok {
			gotApp, err := s.Datastore().GetAppByID(ctx, f.AppID)
			if err != nil {
				handleErrorResponse(c, fmt.Errorf("failed to get app for fn %s", err))
				return
			}
			app = gotApp
			appCache[app.ID] = gotApp
		}

		newF, err := s.fnAnnotator.AnnotateFn(c, app, f)
		if err != nil {
			handleErrorResponse(c, err)
			return
		}
		fns.Items[idx] = newF
	}

	c.JSON(http.StatusOK, fns)
}
