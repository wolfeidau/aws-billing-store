package hive

import (
	"fmt"
	"strings"
)

type HivePartition struct {
	Key   string
	Value string
}

type HivePartitions []HivePartition

// PathString convert a map of partition key/values to a path string
func (hv HivePartitions) PathString() string {

	var parts []string

	for _, v := range hv {
		parts = append(parts, fmt.Sprintf("%s=%s", v.Key, v.Value))
	}

	return strings.Join(parts, "/")
}

// ParsePathString converts a path to a map of partition key/values by locating tokens which contain
// an `=` sign and splitting on that character.
func ParsePathString(path string) HivePartitions {
	parts := strings.Split(path, "/")

	hv := HivePartitions{}

	for _, part := range parts {
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)

			if len(kv) == 2 {
				// hv[kv[0]] = kv[1]
				hv = append(hv, HivePartition{Key: kv[0], Value: kv[1]})
			}
		}
	}

	return hv
}

func (hv HivePartitions) Get(key string) string {
	for _, v := range hv {
		if key == v.Key {
			return v.Value
		}
	}

	return ""
}
