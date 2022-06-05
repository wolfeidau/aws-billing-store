package cur

import (
	"bytes"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	exampleManifest = `{
		"assemblyId": "20220503T120125Z",
		"account": "121212121212",
		"columns": [
			{
				"category": "identity",
				"name": "identity_line_item_id",
				"type": "STRING"
			}
		],
		"charset": "UTF-8",
		"compression": "Parquet",
		"contentType": "Parquet",
		"reportId": "abc123",
		"reportName": "test-managment-cur",
		"billingPeriod": {
			"start": "20220401T000000.000Z",
			"end": "20220501T000000.000Z"
		},
		"bucket": "test-managment-cur",
		"reportKeys": [
			"parquet/test-managment-cur/20220401-20220501/20220503T120125Z/test-managment-cur-00001.snappy.parquet"
		],
		"additionalArtifactKeys": []
	}`
)

func Test_ParseManifestPath(t *testing.T) {
	assert := require.New(t)

	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want *ManifestPeriod
		ok   bool
	}{
		{
			name: "should match period manifest",
			args: args{key: "parquet/test-managment-cur/20220401-20220501/test-managment-cur-Manifest.json"},
			want: &ManifestPeriod{Prefix: "parquet/test-managment-cur", Period: "20220401-20220501"},
			ok:   true,
		},
		{
			name: "should match snapshot manifest",
			args: args{key: "parquet/test-managment-cur/20220401-20220501/20220405T034639Z/test-managment-cur-Manifest.json"},
			want: &ManifestPeriod{Prefix: "parquet/test-managment-cur", Period: "20220401-20220501", Snapshot: "20220405T034639Z"},
			ok:   true,
		},
		{
			name: "should not match manifest",
			args: args{key: "parquet/test-managment-cur/1223/1233/test-managment-cur-Manifest.json"},
			want: nil,
			ok:   false,
		},
		{
			name: "should not match manifest with parquet extension",
			args: args{key: "parquet/test-managment-cur/20220401-20220501/20220503T120125Z/test-managment-cur-00001.snappy.parquet"},
			want: nil,
			ok:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseManifestPath(tt.args.key)
			assert.Equal(tt.want, got)
			assert.Equal(tt.ok, ok)
		})
	}
}

func TestParseManifest(t *testing.T) {
	type args struct {
		rdr io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    *Manifest
		wantErr bool
	}{
		{
			name: "should parse manifest",
			args: args{rdr: bytes.NewBufferString(exampleManifest)},
			want: &Manifest{
				AssemblyID: "20220503T120125Z",
				Account:    "121212121212",
				Columns: []*Column{
					{Category: "identity", Name: "identity_line_item_id", Type: "STRING"},
				},
				Charset:     "UTF-8",
				Compression: "Parquet",
				ContentType: "Parquet",
				ReportID:    "abc123",
				ReportName:  "test-managment-cur",
				BillingPeriod: &BillingPeriod{
					Start: "20220401T000000.000Z",
					End:   "20220501T000000.000Z",
				},
				Bucket: "test-managment-cur",
				ReportKeys: []string{
					"parquet/test-managment-cur/20220401-20220501/20220503T120125Z/test-managment-cur-00001.snappy.parquet",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseManifest(tt.args.rdr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseManifest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBillingPeriod_StartTime(t *testing.T) {
	type fields struct {
		Start string
		End   string
	}
	tests := []struct {
		name    string
		fields  fields
		want    time.Time
		wantErr bool
	}{
		{
			name:    "should parse valid start date",
			fields:  fields{Start: "20220501T000000.000Z"},
			want:    time.Date(2022, 5, 1, 0, 0, 0, 0, time.UTC),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bp := &BillingPeriod{
				Start: tt.fields.Start,
				End:   tt.fields.End,
			}
			got, err := bp.StartTime()
			if (err != nil) != tt.wantErr {
				t.Errorf("BillingPeriod.StartTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BillingPeriod.StartTime() = %v, want %v", got, tt.want)
			}
		})
	}
}
