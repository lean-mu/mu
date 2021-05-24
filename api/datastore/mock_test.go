package datastore

import (
	"testing"

	"github.com/lean-mu/mu/api/datastore/datastoretest"
	"github.com/lean-mu/mu/api/models"
)

func TestDatastore(t *testing.T) {
	f := func(t *testing.T) models.Datastore {
		return NewMock()
	}
	datastoretest.RunAllTests(t, f, datastoretest.NewBasicResourceProvider())
}
