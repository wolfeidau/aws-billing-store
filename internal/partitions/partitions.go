package partitions

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/athena"
	"github.com/aws/aws-sdk-go-v2/service/athena/types"
	"github.com/rs/zerolog/log"
)

var parts = template.Must(template.New("partition").
	Parse(`ALTER TABLE {{ .Table }} ADD IF NOT EXISTS
    PARTITION (year = '{{ .Year }}', month = '{{ .Month }}')`))

type params struct {
	Table string
	Year  string
	Month string
}

type Manager struct {
	athenaClient *athena.Client
	database     string
	table        string
	queryBucket  string
}

func NewManager(queryBucket, region, database, table string) (*Manager, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	athenaClient := athena.NewFromConfig(cfg)

	return &Manager{athenaClient: athenaClient, database: database, table: table, queryBucket: queryBucket}, nil
}

func (m *Manager) CreatePartition(ctx context.Context, year, month string) error {
	query, err := buildQuery(m.table, year, month)
	if err != nil {
		return err
	}

	outputLocation := fmt.Sprintf("s3://%s", m.queryBucket)

	input := &athena.StartQueryExecutionInput{
		QueryString: aws.String(query),
		QueryExecutionContext: &types.QueryExecutionContext{
			Database: aws.String(m.database),
		},
		ResultConfiguration: &types.ResultConfiguration{
			OutputLocation: aws.String(outputLocation),
		},
	}

	log.Ctx(ctx).Info().Str("query", query).Str("outputLocation", outputLocation).Msg("run query")

	result, err := m.athenaClient.StartQueryExecution(ctx, input)
	if err != nil {
		return err
	}

	// Wait for query execution to complete
	for {
		status, err := m.athenaClient.GetQueryExecution(ctx, &athena.GetQueryExecutionInput{
			QueryExecutionId: result.QueryExecutionId,
		})
		if err != nil {
			return err
		}

		switch status.QueryExecution.Status.State {
		case types.QueryExecutionStateSucceeded:
			log.Ctx(ctx).Info().Str("QueryExecutionId", aws.ToString(status.QueryExecution.QueryExecutionId)).Msg("query execution succeeded")
			return nil
		case types.QueryExecutionStateCancelled:
			log.Ctx(ctx).Info().Str("QueryExecutionId", aws.ToString(status.QueryExecution.QueryExecutionId)).Msg("query execution cancelled")
			return fmt.Errorf("Athena query execution cancelled")
		case types.QueryExecutionStateFailed:
			return fmt.Errorf("Athena query execution failed: %s", aws.ToString(status.QueryExecution.Status.AthenaError.ErrorMessage))
		}

		select {
		case <-time.After(time.Second):
			// Time to retry
		case <-ctx.Done():
			// If the context was cancelled, cancel the running query
			return ctx.Err()
		}
	}
}

func buildQuery(table string, year, month string) (string, error) {
	buf := new(bytes.Buffer)

	err := parts.Execute(buf, &params{Table: table, Year: year, Month: month})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
