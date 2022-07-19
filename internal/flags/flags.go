package flags

import "github.com/alecthomas/kong"

type Base struct {
	Version         kong.VersionFlag
	RawEventLogging bool   `help:"Enable raw event logging." env:"RAW_EVENT_LOGGING"`
	Debug           bool   `help:"Enable debug logging." env:"DEBUG"`
	Stage           string `help:"The development stage." env:"STAGE"`
	Branch          string `help:"The git branch this code originated." env:"BRANCH"`
}

type Symlink struct {
	Base
}

type Partitions struct {
	Base
	Region      string `help:"The AWS region." env:"AWS_REGION"`
	QueryBucket string `help:"The Bucket used for Athena query results." env:"QUERY_BUCKET"`
	Database    string `help:"The Athena database name." env:"DATABASE"`
	Table       string `help:"The Athena table name." env:"TABLE"`
}
