package hive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePathString(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want HivePartitions
	}{
		{
			name: "should parse valid path",
			args: args{path: "hive/year=2020/month=1/symlink.txt"},
			want: HivePartitions{
				{Key: "year", Value: "2020"},
				{Key: "month", Value: "1"},
			},
		},
		{
			name: "should parse valid path",
			args: args{path: "hive/year=2020/month=1/symlink.txt"},
			want: HivePartitions{
				{Key: "year", Value: "2020"},
				{Key: "month", Value: "1"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePathString(tt.args.path)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHivePartitions_PathString(t *testing.T) {
	tests := []struct {
		name string
		hv   HivePartitions
		want string
	}{
		{
			name: "should build valid path",
			hv: HivePartitions{
				{Key: "year", Value: "2022"},
				{Key: "month", Value: "1"},
			},
			want: "year=2022/month=1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.hv.PathString(); got != tt.want {
				t.Errorf("HivePartitions.PathString() = %v, want %v", got, tt.want)
			}
		})
	}
}
