package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"

	"pedro.to/rcaptv/auth"
	"pedro.to/rcaptv/auth/pb"
	"pedro.to/rcaptv/certs"
	cfg "pedro.to/rcaptv/config"
	"pedro.to/rcaptv/database"
	"pedro.to/rcaptv/database/postgres"
	"pedro.to/rcaptv/utils"
)

var ErrPingTimeout = errors.New("ping timed out")

type RPCTokenValidator struct {
	pb.UnimplementedTokenValidatorServiceServer
	tv *auth.TokenValidator
}

func (v *RPCTokenValidator) AddUser(ctx context.Context, u *pb.User) (*pb.AddUserReply, error) {
	v.tv.AddUser(u.Id)
	return new(pb.AddUserReply), nil
}

func (v *RPCTokenValidator) Ping(ctx context.Context, empty *emptypb.Empty) (*wrapperspb.BoolValue, error) {
	for {
		select {
		case <-ctx.Done():
			return wrapperspb.Bool(false), ErrPingTimeout
		case <-v.tv.Ready():
			return wrapperspb.Bool(true), nil
		}
	}
}

func NewRPCServer(tv *auth.TokenValidator) *grpc.Server {
	sv := &RPCTokenValidator{
		tv: tv,
	}
	creds, err := credentials.NewServerTLSFromFile(
		certs.Filename("server_cert.pem"),
		certs.Filename("server_key.pem"),
	)
	if err != nil {
		panic("failed to construct RPC credentials: " + err.Error())
	}
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
	)
	pb.RegisterTokenValidatorServiceServer(grpcServer, sv)
	return grpcServer
}

func main() {
	l := log.With().Str("ctx", "auth_service").Logger()
	l.Info().Msgf("starting auth service (v%s)", cfg.Version)
	if !cfg.IsProd {
		l.Warn().Msg("[!] running auth service in dev mode")
	}
	sto := database.New(postgres.New(
		&database.StorageOptions{
			StorageHost:     cfg.PostgresHost,
			StoragePort:     cfg.PostgresPort,
			StorageUser:     cfg.RcaptvPostgresUser,
			StoragePassword: cfg.RcaptvPostgresPassword,
			StorageDbName:   cfg.PostgresDBName,

			StorageMaxIdleConns:    cfg.PostgresMaxIdleConns,
			StorageMaxOpenConns:    cfg.PostgresMaxOpenConns,
			StorageConnMaxLifetime: time.Duration(cfg.PostgresConnMaxLifetimeMinutes) * time.Minute,
			StorageConnTimeout:     time.Duration(cfg.PostgresConnTimeoutSeconds) * time.Second,

			MigrationVersion: cfg.PostgresMigVersion,
			MigrationPath:    cfg.PostgresMigPath,
		}))
	svc := auth.NewService(sto)
	svc.Start()
	grcpServer := NewRPCServer(svc.TokenValidator)
	listener, err := net.Listen("tcp",
		fmt.Sprintf(":%s", cfg.RPCAuthPort),
	)
	if err != nil {
		panic("failed to create RPC TCP listener: " + err.Error())
	}
	go func() {
		l.Info().Msgf("gRPC server listening on %s", listener.Addr())
		if err := grcpServer.Serve(listener); err != nil {
			panic("grpc server: " + err.Error())
		}
	}()

	sig := utils.WaitInterrupt()
	l.Info().Msgf("termination signal received [%s]. Attempting to gracefully shutdown...", sig)
	l.Info().Msg("stopping auth service")
	grcpServer.GracefulStop()
	svc.Stop()
}

func init() {
	cfg.Setup()
}
