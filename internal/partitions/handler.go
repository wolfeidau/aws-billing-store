package partitions

import (
	"context"
	"errors"
	"fmt"

	"github.com/wolfeidau/aws-billing-store/internal/events"
	"github.com/wolfeidau/aws-billing-store/internal/events/s3created"
	"github.com/wolfeidau/aws-billing-store/internal/flags"
	"github.com/wolfeidau/aws-billing-store/internal/hive"
)

type Handler struct {
	cfg  flags.Partitions
	pman *Manager
}

func NewHandler(ctx context.Context, cfg flags.Partitions) (*Handler, error) {

	pman, err := NewManager(cfg.QueryBucket, cfg.Region, cfg.Database, cfg.Table)
	if err != nil {
		return nil, err
	}

	return &Handler{
		cfg:  cfg,
		pman: pman,
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

	hivePartitions := hive.ParsePathString(created.Object.Key)

	if hivePartitions["year"] == "" || hivePartitions["month"] == "" {
		return nil, fmt.Errorf("unable to parse hive partitions from path: %s", created.Object.Key)
	}

	err := h.pman.CreatePartition(ctx, hivePartitions["year"], hivePartitions["month"])
	if err != nil {
		return nil, err
	}

	return []byte(`{"msg": "ok"}`), nil
}
