package main

import (
	"context"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"
	"github.com/wolfeidau/aws-billing-store/internal/flags"
	"github.com/wolfeidau/aws-billing-store/internal/partitions"
	lmw "github.com/wolfeidau/lambda-go-extras/middleware"
	"github.com/wolfeidau/lambda-go-extras/standard"
)

var (
	// assigned during build time with -ldflags
	commit = "unknown"

	cli flags.Partition
)

func main() {
	kong.Parse(&cli,
		kong.Vars{"version": commit}, // bind a var for version
	)

	flds := lmw.FieldMap{"commit": commit}

	pman, err := partitions.NewManager(cli.QueryBucket, cli.Region, cli.Database, cli.Table)
	if err != nil {
		log.Fatal().Err(err).Msg("partitions manager failed")
	}

	h, err := partitions.NewHandler(context.Background(), pman)
	if err != nil {
		log.Fatal().Err(err).Msg("handler setup failed")
	}

	standard.GenericDefault(h.Handler, standard.Fields(flds))
}
