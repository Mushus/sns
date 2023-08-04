//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Mushus/activitypub"
	"github.com/Mushus/activitypub/sqlite"
	"github.com/google/wire"
	"github.com/rs/zerolog"
)

func createServer(log *zerolog.Logger) (*activitypub.Server, error) {
	wire.Build(
		activitypub.NewHandler,
		activitypub.NewServer,
		activitypub.NewURLResolver,
		activitypub.ParseConfig,
		activitypub.NewProcessor,
		activitypub.NewRemoteServer,
		sqlite.NewSession,
		sqlite.NewSQLite,
		sqlite.NewAccountDB,
		sqlite.NewFollowDB,
	)
	return nil, nil
}
