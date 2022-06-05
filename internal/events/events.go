package events

import (
	"encoding/json"
	"time"

	"github.com/wolfeidau/aws-billing-service/internal/events/s3created"
)

type AWSEvent struct {
	Account    string          `json:"account"`
	RawDetail  json.RawMessage `json:"detail"`
	Detail     interface{}     `json:"-"`
	DetailType string          `json:"detail-type"`
	Id         string          `json:"id"`
	Region     string          `json:"region"`
	Resources  []string        `json:"resources"`
	Source     string          `json:"source"`
	Time       time.Time       `json:"time"`
	Version    string          `json:"version"`
}

func (ae *AWSEvent) Matches(source, detailType string) bool {
	return ae.Source == source && ae.DetailType == detailType
}

func (ae *AWSEvent) UpdateDetail(v interface{}) error {

	err := json.Unmarshal(ae.RawDetail, v)
	if err != nil {
		return err
	}

	ae.Detail = v

	return nil
}

func ParseEvent(payload []byte) (*AWSEvent, error) {
	event := new(AWSEvent)

	err := json.Unmarshal(payload, event)
	if err != nil {
		return nil, err
	}

	switch {
	case event.Matches("aws.s3", "Object Created"):
		objectCreate := new(s3created.ObjectCreated)
		err = event.UpdateDetail(objectCreate)
	}
	if err != nil {
		return nil, err
	}

	return event, nil
}
