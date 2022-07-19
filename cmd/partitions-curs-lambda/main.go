package main

import (
	"context"
	"runtime/debug"

	"github.com/alecthomas/kong"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/aws-billing-store/internal/flags"
	"github.com/wolfeidau/aws-billing-store/internal/partitions"
	lmw "github.com/wolfeidau/lambda-go-extras/middleware"
	"github.com/wolfeidau/lambda-go-extras/middleware/raw"
	zlog "github.com/wolfeidau/lambda-go-extras/middleware/zerolog"
)

var (
	// assigned during build time with -ldflags
	commit = "unknown"

	cli flags.Partitions
)

func main() {
	kong.Parse(&cli,
		kong.Vars{"version": commit}, // bind a var for version
	)

	flds := lmw.FieldMap{"commit": commit}

	if buildInfo, ok := debug.ReadBuildInfo(); ok {
		log.Info().Fields(map[string]interface{}{
			"buildInfo": buildInfo,
		}).Msg("startup")
	}

	h, err := partitions.NewHandler(context.Background(), cli)
	if err != nil {
		log.Fatal().Err(err).Msg("handler setup failed")
	}

	ch := lmw.New(
		raw.New(raw.Fields(flds)),   // raw event logger primarily used during development
		zlog.New(zlog.Fields(flds)), // inject zerolog into the context
	).Then(h)

	lambda.Start(ch)
}
