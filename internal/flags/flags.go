package flags

import "github.com/alecthomas/kong"

type S3Events struct {
	Version         kong.VersionFlag
	RawEventLogging bool   `help:"Enable raw event logging." env:"RAW_EVENT_LOGGING"`
	Debug           bool   `help:"Enable debug logging." env:"DEBUG"`
	Stage           string `help:"The development stage." env:"STAGE"`
	Branch          string `help:"The git branch this code originated." env:"BRANCH"`
}
