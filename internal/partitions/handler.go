package partitions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchevents/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/aws-billing-store/internal/cur"
	"github.com/wolfeidau/aws-billing-store/internal/events"
	"github.com/wolfeidau/aws-billing-store/internal/events/s3created"
	"github.com/wolfeidau/aws-billing-store/internal/hive"
)

type partitionEvent struct {
	Account        string              `json:"account"`
	AssemblyID     string              `json:"assembly_id"`
	BillingPeriod  *cur.BillingPeriod  `json:"billing_period"`
	HivePartitions hive.HivePartitions `json:"hive_partitions"`
	ReportKeys     []string            `json:"report_keys"`
	Manifest       string              `json:"manifest"`
	Symlink        string              `json:"symlink"`
}

type Handler struct {
	s3client     *s3.Client
	eventsclient *cloudwatchevents.Client
	sgen         *hive.SymlinkGenerator
	pman         *Manager
}

func NewHandler(ctx context.Context, pman *Manager) (*Handler, error) {
	config, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	s3client := s3.NewFromConfig(config)
	eventsclient := cloudwatchevents.NewFromConfig(config)

	return &Handler{
		s3client:     s3client,
		eventsclient: eventsclient,
		sgen:         hive.NewSymlinkGenerator(s3client),
		pman:         pman,
	}, nil
}

func (h *Handler) Invoke(ctx context.Context, payload []byte) ([]byte, error) {

	event, err := events.ParseEvent(payload)
	if err != nil {
		return nil, err
	}

	switch v := event.Detail.(type) {
	case *s3created.ObjectCreated:
		return h.processCreated(ctx, v)
	}

	return nil, errors.New("failed to process event, unknown type")
}

func (h *Handler) processCreated(ctx context.Context, created *s3created.ObjectCreated) ([]byte, error) {

	// does the key match the manifest path structure
	manifestPeriod, ok := cur.ParseManifestPath(created.Object.Key)
	if !ok {
		log.Ctx(ctx).Info().Str("object", created.Object.Key).Msg("skipped file as it is not a manifest")
		return []byte(`{"msg": "skipped"}`), nil
	}

	// at the moment we skip snapshot manifests
	if manifestPeriod.Snapshot != "" {
		log.Ctx(ctx).Info().Str("object", created.Object.Key).Msg("skipped file as it is a snapshot manifest")
		return []byte(`{"msg": "skipped"}`), nil
	}

	res, err := h.s3client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(created.Bucket.Name),
		Key:    aws.String(created.Object.Key),
	})
	if err != nil {
		return nil, err
	}

	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(res.Body)

	manifest, err := cur.ParseManifest(res.Body)
	if err != nil {
		return nil, err
	}

	log.Ctx(ctx).Info().Str("AssemblyID", manifest.AssemblyID).Msg("loaded manifest")

	startDate, err := manifest.BillingPeriod.StartTime()
	if err != nil {
		return nil, err
	}

	year := fmt.Sprintf("%d", startDate.Year())
	month := fmt.Sprintf("%d", startDate.Month())

	hivePartitions := hive.HivePartitions{
		{Key: "year", Value: year},
		{Key: "month", Value: month},
	}

	keys := make([]string, len(manifest.ReportKeys))

	for i, reportKey := range manifest.ReportKeys {
		keys[i] = fmt.Sprintf("s3://%s/%s", created.Bucket.Name, reportKey)
	}

	storeRes, err := h.sgen.StoreSymlink(ctx, created.Bucket.Name, manifestPeriod.Prefix, hivePartitions, keys)
	if err != nil {
		return nil, err
	}

	err = h.pman.CreatePartition(ctx, year, month)
	if err != nil {
		return nil, err
	}

	pe := partitionEvent{
		HivePartitions: hivePartitions,
		Account:        manifest.Account,
		AssemblyID:     manifest.AssemblyID,
		BillingPeriod:  manifest.BillingPeriod,
		ReportKeys:     keys,
		Manifest:       fmt.Sprintf("s3://%s/%s", created.Bucket.Name, created.Object.Key),
		Symlink:        fmt.Sprintf("s3://%s/%s", storeRes.Bucket, storeRes.Key),
	}

	log.Ctx(ctx).Info().Fields(map[string]any{"partitionEvent": pe}).Msg("publish event")

	data, err := json.Marshal(pe)
	if err != nil {
		return nil, err
	}

	_, err = h.eventsclient.PutEvents(ctx, &cloudwatchevents.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{
			{
				Detail:     aws.String(string(data)),
				DetailType: aws.String("Partition Created"),
				Resources:  []string{manifest.AssemblyID},
				Source:     aws.String("aws-billing-store"),
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return []byte(`{"msg": "ok"}`), nil
}
