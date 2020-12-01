package anser

import (
	"context"

	"github.com/cdr/amboy"
	"github.com/cdr/amboy/job"
	"github.com/cdr/amboy/registry"
	"github.com/deciduosity/anser/model"
	"github.com/cdr/grip"
	"github.com/cdr/grip/message"
	"github.com/pkg/errors"
)

func init() {
	registry.AddJobType("stream-migration", func() amboy.Job { return makeStreamProducer() })
}

func NewStreamMigration(e Environment, m model.Stream) Migration {
	j := makeStreamProducer()
	j.Definition = m
	j.MigrationHelper = NewMigrationHelper(e)
	return j
}

func makeStreamProducer() *streamMigrationJob {
	return &streamMigrationJob{
		MigrationHelper: &migrationBase{},
		Base: job.Base{
			JobType: amboy.JobType{
				Name:    "stream-migration",
				Version: 0,
			},
		},
	}
}

type streamMigrationJob struct {
	Definition      model.Stream `bson:"migration" json:"migration" yaml:"migration"`
	job.Base        `bson:"job_base" json:"job_base" yaml:"job_base"`
	MigrationHelper `bson:"-" json:"-" yaml:"-"`
}

func (j *streamMigrationJob) Run(ctx context.Context) {
	grip.Info(message.Fields{
		"message":   "starting migration",
		"migration": j.Definition.Migration,
		"operation": "stream",
		"id":        j.ID(),
		"ns":        j.Definition.Namespace,
		"name":      j.Definition.ProcessorName,
	})

	defer j.FinishMigration(ctx, j.Definition.Migration, &j.Base)

	env := j.Env()

	if producer, ok := env.GetDocumentProcessor(j.Definition.ProcessorName); ok {
		client, err := env.GetClient()
		if err != nil {
			j.AddError(errors.Wrap(err, "problem getting database client"))
			return
		}

		iter := producer.Load(client, j.Definition.Namespace, j.Definition.Query)
		if iter == nil {
			j.AddError(errors.Errorf("document processor for %s could not return iterator",
				j.Definition.Migration))
			return
		}

		j.AddError(producer.Migrate(iter))
	} else {
		j.AddError(errors.Errorf("producer named '%s' is not defined", j.Definition.ProcessorName))
	}

}
