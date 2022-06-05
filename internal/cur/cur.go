package cur

import (
	"encoding/json"
	"io"
	"regexp"
)

var (
	manifestRegex = regexp.MustCompile(`(.*)\/(\d{8}-\d{8})\/(\d{8}T\d{6}Z)?.*Manifest.json$`)
)

type Column struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Type     string `json:"type"`
}

type BillingPeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type Manifest struct {
	AssemblyID    string        `json:"assemblyId"`
	Account       string        `json:"account"`
	Columns       []Column      `json:"columns"`
	Charset       string        `json:"charset"`
	Compression   string        `json:"compression"`
	ContentType   string        `json:"contentType"`
	ReportID      string        `json:"reportId"`
	ReportName    string        `json:"reportName"`
	BillingPeriod BillingPeriod `json:"billingPeriod"`
	Bucket        string        `json:"bucket"`
	ReportKeys    []string      `json:"reportKeys"`
}

func ParseManifest(rdr io.Reader) (*Manifest, error) {
	manifest := new(Manifest)

	dec := json.NewDecoder(rdr)

	err := dec.Decode(manifest)
	if err != nil {
		return nil, err
	}

	return manifest, nil
}

type ManifestPeriod struct {
	Prefix   string
	Period   string
	Snapshot string
}

func ParseManifestPath(key string) (*ManifestPeriod, bool) {
	res := manifestRegex.FindAllStringSubmatch(key, -1)
	for i := range res {
		return &ManifestPeriod{Prefix: res[i][1], Period: res[i][2], Snapshot: res[i][3]}, true
	}

	return nil, false
}
