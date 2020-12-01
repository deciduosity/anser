package mock

import (
	"github.com/cdr/amboy"
	"github.com/cdr/amboy/dependency"
	"github.com/deciduosity/anser/client"
	"github.com/deciduosity/anser/model"
	"github.com/cdr/grip"
	"github.com/pkg/errors"
)

type Environment struct {
	ShouldPreferClient bool
	IsSetup            bool
	ReturnNilClient    bool
	SetupError         error
	ClientError        error
	QueueError         error
	NetworkError       error

	PreferedSetup      interface{}
	Client             *Client
	Network            *DependencyNetwork
	Queue              amboy.Queue
	SetupClient        client.Client
	Closers            []func() error
	DependencyManagers map[string]*DependencyManager
	MigrationRegistry  map[string]client.MigrationOperation
	ProcessorRegistry  map[string]client.Processor
	MetaNS             model.Namespace
}

func NewEnvironment() *Environment {
	return &Environment{
		Client:             NewClient(),
		Network:            NewDependencyNetwork(),
		DependencyManagers: make(map[string]*DependencyManager),
		MigrationRegistry:  make(map[string]client.MigrationOperation),
		ProcessorRegistry:  make(map[string]client.Processor),
	}
}

func (e *Environment) Setup(q amboy.Queue, cl client.Client, name string) error {
	e.Queue = q
	e.IsSetup = true
	e.SetupClient = cl
	e.MetaNS.DB = name
	return e.SetupError
}

func (e *Environment) GetClient() (client.Client, error) {
	if e.ClientError != nil {
		return nil, e.ClientError
	}

	if e.ReturnNilClient {
		return nil, nil
	}

	return e.Client, nil

}

func (e *Environment) GetQueue() (amboy.Queue, error) {
	if e.QueueError != nil {
		return nil, e.QueueError
	}

	return e.Queue, nil
}

func (e *Environment) GetDependencyNetwork() (model.DependencyNetworker, error) {
	if e.NetworkError != nil {
		return nil, e.NetworkError
	}

	return e.Network, nil
}

func (e *Environment) RegisterManualMigrationOperation(name string, op client.MigrationOperation) error {
	if _, ok := e.MigrationRegistry[name]; ok {
		return errors.Errorf("migration operation %s already exists", name)
	}

	e.MigrationRegistry[name] = op
	return nil
}

func (e *Environment) GetManualMigrationOperation(name string) (client.MigrationOperation, bool) {
	op, ok := e.MigrationRegistry[name]
	return op, ok
}

func (e *Environment) RegisterDocumentProcessor(name string, docp client.Processor) error {
	if _, ok := e.ProcessorRegistry[name]; ok {
		return errors.Errorf("document processor named %s already registered", name)
	}

	e.ProcessorRegistry[name] = docp
	return nil
}

func (e *Environment) GetDocumentProcessor(name string) (client.Processor, bool) {
	docp, ok := e.ProcessorRegistry[name]
	return docp, ok
}

func (e *Environment) MetadataNamespace() model.Namespace { return e.MetaNS }

func (e *Environment) NewDependencyManager(n string) dependency.Manager {
	edges := dependency.NewJobEdges()
	e.DependencyManagers[n] = &DependencyManager{
		Name:     n,
		JobEdges: &edges,
	}

	return e.DependencyManagers[n]
}

func (e *Environment) RegisterCloser(closer func() error) { e.Closers = append(e.Closers, closer) }
func (e *Environment) Close() error {
	catcher := grip.NewSimpleCatcher()
	for _, closer := range e.Closers {
		catcher.Add(closer())
	}
	return catcher.Resolve()
}
