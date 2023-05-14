package main

import (
	"context"
	"database/sql"
	"github.com/gaggudeep/bank_go/api"
	"github.com/gaggudeep/bank_go/const/enum"
	db "github.com/gaggudeep/bank_go/db/sqlc"
	_ "github.com/gaggudeep/bank_go/doc/statik"
	"github.com/gaggudeep/bank_go/gapi"
	"github.com/gaggudeep/bank_go/pb"
	"github.com/gaggudeep/bank_go/util"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	_ "github.com/lib/pq"
	"github.com/rakyll/statik/fs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/encoding/protojson"
	"net"
	"net/http"
	"os"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal().Msgf("cannot load config: %s", err)
	}

	if config.Environment == enum.EnvDev {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	dbConn, err := sql.Open(config.DBDriver, config.DBURL)
	if err != nil {
		log.Fatal().Msgf("cannot connect to db: %s", err)
	}

	runDBMigration(config.MigrationURL, config.DBURL)

	store := db.NewStore(dbConn)
	go runGatewayServer(&config, store)
	runGRPCServer(&config, store)
}

func runDBMigration(migrationURL string, dbURL string) {
	migration, err := migrate.New(migrationURL, dbURL)
	if err != nil {
		log.Fatal().Msgf("cannot create new migration instance: %s", err)
	}
	if err = migration.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal().Msgf("failed to run migrate up: %s", err)
	}

	log.Info().Msg("db migrated successfully")
}

func runGinServer(config *util.Config, store db.Store) {
	server, err := api.NewServer(store, config)
	if err != nil {
		log.Fatal().Msgf("cannot create server: %s", err)
	}

	err = server.Start(config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msgf("cannot start server: %s", err)
	}
}

func runGRPCServer(config *util.Config, store db.Store) {
	server, err := gapi.NewServer(store, config)
	if err != nil {
		log.Fatal().Msgf("cannot create server: %s", err)
	}

	gRPCLogger := grpc.UnaryInterceptor(gapi.GRPCLogger)
	gRPCServer := grpc.NewServer(gRPCLogger)

	pb.RegisterBankServer(gRPCServer, server)
	reflection.Register(gRPCServer)

	listener, err := net.Listen("tcp", config.GRPCServerAddress)
	if err != nil {
		log.Fatal().Msgf("cannot create listener: %s", err)
	}

	log.Info().Msgf("starting gRPC server at %s", listener.Addr().String())
	err = gRPCServer.Serve(listener)
	if err != nil {
		log.Fatal().Msgf("cannot start grpc server: %s", err)
	}
}

func runGatewayServer(config *util.Config, store db.Store) {
	server, err := gapi.NewServer(store, config)
	if err != nil {
		log.Fatal().Msgf("cannot create server: %s", err)
	}

	grpcMux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				UseProtoNames: true,
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = pb.RegisterBankHandlerServer(ctx, grpcMux, server)
	if err != nil {
		log.Fatal().Msgf("cannot register handler server: %s", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", grpcMux)

	statikFS, err := fs.New()
	if err != nil {
		log.Fatal().Msgf("cannot create statik fs: %s", err)
	}
	swaggerHandler := http.StripPrefix("/swagger/", http.FileServer(statikFS))
	mux.Handle("/swagger/", swaggerHandler)

	listener, err := net.Listen("tcp", config.HTTPServerAddress)
	if err != nil {
		log.Fatal().Msgf("cannot create listener: %s", err)
	}

	log.Info().Msgf("starting http gateway server at %s", listener.Addr().String())

	handler := gapi.HTTPLogger(mux)
	err = http.Serve(listener, handler)
	if err != nil {
		log.Fatal().Msgf("cannot start http gateway server: %s", err)
	}
}
