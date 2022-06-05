package main

import (
	"context"
	"runtime/debug"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	lmw "github.com/wolfeidau/lambda-go-extras/middleware"
	"github.com/wolfeidau/lambda-go-extras/middleware/raw"
	zlog "github.com/wolfeidau/lambda-go-extras/middleware/zerolog"
)

// assigned during build time with -ldflags
var (
	commit = "unknown"
)

func main() {

	flds := lmw.FieldMap{"commit": commit}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		log.Info().Fields(map[string]interface{}{
			"buildInfo": buildInfo,
		}).Msg("startup")
	}

	ch := lmw.New(
		raw.New(raw.Fields(flds)),   // raw event logger primarily used during development
		zlog.New(zlog.Fields(flds)), // inject zerolog into the context
	).ThenFunc(processEvent)

	lambda.StartHandler(ch)
}

func processEvent(ctx context.Context, payload []byte) ([]byte, error) {
	zerolog.Ctx(ctx).Info().Msg("processEvent")
	return []byte("ok"), nil
}
