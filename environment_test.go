package anser

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cdr/amboy"
	"github.com/cdr/amboy/queue"
	"github.com/deciduosity/anser/client"
	"github.com/deciduosity/anser/db"
	"github.com/cdr/grip"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	grip.SetName("anser.test")
}

type EnvImplSuite struct {
	env     *envState
	q       amboy.Queue
	client  client.Client
	session db.Session
	cancel  context.CancelFunc
	suite.Suite
}

func TestEnvImplSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(EnvImplSuite))
}

func (s *EnvImplSuite) SetupTest() {
	ctx, cancel := context.WithCancel(context.Background())
	s.q = queue.NewLocalLimitedSize(4, 256)
	s.cancel = cancel
	s.NoError(s.q.Start(ctx))

	s.Require().Equal(globalEnv, GetEnvironment())

	cl, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017").SetConnectTimeout(10 * time.Millisecond))
	s.Require().NoError(err)
	s.client = client.WrapClient(cl)

	s.env = &envState{
		migrations: make(map[string]client.MigrationOperation),
		processor:  make(map[string]client.Processor),
	}

	s.False(s.env.isSetup)
	s.NoError(s.env.Setup(s.q, s.client, "anser"))
	s.True(s.env.isSetup)
	s.Equal(s.env.metadataNS.DB, defaultAnserDB)
	s.Equal(s.env.metadataNS.Collection, defaultMetadataCollection)
	s.Equal(s.env.MetadataNamespace(), s.env.metadataNS)
}

func (s *EnvImplSuite) TearDownTest() {
	s.cancel()
}

func (s *EnvImplSuite) TestCallingSetupMultipleTimesErrors() {
	s.Error(s.env.Setup(s.q, s.client, "anser_test"))
	s.True(s.env.isSetup)
}

func (s *EnvImplSuite) TestUnstartedQueueCausesError() {
	s.env.isSetup = false
	s.env.queue = nil

	s.Error(s.env.Setup(queue.NewLocalLimitedSize(2, 256), s.client, ""))
	s.Nil(s.env.queue)
	s.False(s.env.isSetup)
}

func (s *EnvImplSuite) TestQueueAccessor() {
	queue, err := s.env.GetQueue()
	s.NoError(err)
	s.Equal(s.q, queue)
}

func (s *EnvImplSuite) TestDepNetworkAccessor() {
	network, err := s.env.GetDependencyNetwork()
	s.NoError(err)
	s.NotNil(network)
	s.NotEqual(globalEnv.deps, network)
	s.Equal(s.env.deps, network)
}

func (s *EnvImplSuite) TestDependencyNetworkConstructor() {
	dep := s.env.NewDependencyManager("foo")

	s.NotNil(dep)
	mdep := dep.(*migrationDependency)
	s.Equal(mdep.Env(), s.env)
	s.Equal(mdep.MigrationID, "foo")
}

func (s *EnvImplSuite) TestRegisterCloser() {
	s.Len(s.env.closers, 0)
	s.env.RegisterCloser(nil)
	s.Len(s.env.closers, 0)
	s.env.RegisterCloser(func() error { return nil })
	s.Len(s.env.closers, 1)

	s.env.RegisterCloser(func() error { return nil })
	s.Len(s.env.closers, 2)

	s.NoError(s.env.Close())

	s.env.RegisterCloser(func() error { return errors.New("foo") })
	s.Len(s.env.closers, 3)

	s.Error(s.env.Close())
}

func (s *EnvImplSuite) TestCloseEncountersError() {
	s.Len(s.env.closers, 0)
	s.env.RegisterCloser(func() error { return errors.New("foo") })
	s.Len(s.env.closers, 1)

	s.Error(s.env.Close())
}

func TestUninitializedEnvErrors(t *testing.T) {
	t.Parallel()
	assert := assert.New(t)
	env := &envState{}

	queue, err := env.GetQueue()
	assert.Nil(queue)
	assert.Error(err)

	net, err := env.GetDependencyNetwork()
	assert.Nil(net)
	assert.Error(err)
}
