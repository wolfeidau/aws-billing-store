package partitions

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"os"
	"text/template"

	"github.com/rs/zerolog/log"
	drv "github.com/uber/athenadriver/go"
)

var parts = template.Must(template.New("partition").
	Parse(`ALTER TABLE {{ .Table }} ADD IF NOT EXISTS
PARTITION (year=?, month=?)`))

type params struct {
	Table string
}

type Manager struct {
	conf     *drv.Config
	database string
	table    string
}

func NewManager(queryBucket, region, database, table string) (*Manager, error) {
	err := os.Setenv("AWS_SDK_LOAD_CONFIG", "1")
	if err != nil {
		return nil, err
	}

	conf, err := drv.NewDefaultConfig(
		fmt.Sprintf("s3://%s/results", queryBucket),
		region,
		drv.DummyAccessID,
		drv.DummySecretAccessKey,
	)
	if err != nil {
		return nil, err
	}

	conf.SetDB(database)

	return &Manager{conf: conf, database: database, table: table}, nil
}

func (m *Manager) CreatePartition(ctx context.Context, year, month string) error {
	query, err := buildQuery(m.table)
	if err != nil {
		return err
	}

	db, err := sql.Open(drv.DriverName, m.conf.Stringify())
	if err != nil {
		return err
	}

	log.Ctx(ctx).Info().Str("query", query).Msg("run query")

	_, err = db.Query(query, year, month)
	if err != nil {
		return err
	}

	return nil
}

func buildQuery(table string) (string, error) {
	buf := new(bytes.Buffer)

	err := parts.Execute(buf, &params{Table: table})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
