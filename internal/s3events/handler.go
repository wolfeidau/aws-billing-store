package s3events

import (
	"context"
	"errors"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/aws-billing-service/internal/cur"
	"github.com/wolfeidau/aws-billing-service/internal/events"
	"github.com/wolfeidau/aws-billing-service/internal/events/s3created"
	"github.com/wolfeidau/aws-billing-service/internal/flags"
)

type Handler struct {
	cfg      flags.S3Events
	s3client *s3.Client
}

func NewHandler(ctx context.Context, cfg flags.S3Events) (*Handler, error) {
	config, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	s3client := s3.NewFromConfig(config)

	return &Handler{
		cfg:      cfg,
		s3client: s3client,
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
	if manifestPeriod.Snapshot == "" {
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

	return []byte(`{"msg": "ok"}`), nil
}
