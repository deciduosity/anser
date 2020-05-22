package mock

import (
	"testing"

	"github.com/deciduosity/amboy/dependency"
	"github.com/deciduosity/anser"
	"github.com/deciduosity/anser/client"
	"github.com/deciduosity/anser/db"
	"github.com/deciduosity/anser/model"
	"github.com/stretchr/testify/assert"
)

func TestInterfaces(t *testing.T) {
	assert := assert.New(t)

	assert.Implements((*db.Session)(nil), &Session{})
	assert.Implements((*db.Database)(nil), &LegacyDatabase{})
	assert.Implements((*db.Collection)(nil), &LegacyCollection{})
	assert.Implements((*db.Query)(nil), &Query{})
	assert.Implements((*db.Results)(nil), &Query{})
	assert.Implements((*db.Results)(nil), &Pipeline{})
	assert.Implements((*db.Iterator)(nil), &Iterator{})

	assert.Implements((*model.DependencyNetworker)(nil), &DependencyNetwork{})
	assert.Implements((*anser.Environment)(nil), &Environment{})
	assert.Implements((*dependency.Manager)(nil), &DependencyManager{})

	assert.Implements((*client.Client)(nil), &Client{})
	assert.Implements((*client.Database)(nil), &Database{})
	assert.Implements((*client.Collection)(nil), &Collection{})
}
