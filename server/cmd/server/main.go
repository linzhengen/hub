package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/soheilhy/cmux"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/linzhengen/hub/v1/server/config"
	"github.com/linzhengen/hub/v1/server/di"
	"github.com/linzhengen/hub/v1/server/internal/infrastructure/persistence/mysql"
	httphandler "github.com/linzhengen/hub/v1/server/internal/interface/http"
	"github.com/linzhengen/hub/v1/server/internal/usecase/develop"

	"github.com/linzhengen/hub/v1/server/pkg/logger"
)

func main() {
	if err := serverCmd.Execute(); err != nil {
		logger.Severef("failed execute, err: %v", err)
	}
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "hub server",
	Run: func(cmd *cobra.Command, args []string) {
		svr := &server{}
		if err := svr.run(cmd.Context(), config.New(cmd.Context())); err != nil {
			logger.Severef("failed to run server, err: %v", err)
		}
	},
}

type server struct {
	grpcServer     *grpc.Server
	httpServeMux   *runtime.ServeMux
	migrateUseCase develop.MigrateUseCase
	seedUseCase    develop.SeedUseCase
}

func (s *server) run(ctx context.Context, envCfg config.EnvConfig) error {
	// Start listener
	var lis net.Listener
	var listerErr error
	address := fmt.Sprintf(":%d", envCfg.Port)
	lis, listerErr = net.Listen("tcp", address)

	if listerErr != nil {
		logger.Severef("failed to listen: %v", listerErr)
		return listerErr
	}

	// Cmux is used to support servicing gRPC and HTTP1.1+JSON on the same port
	tcpm := cmux.New(lis)
	httpL := tcpm.Match(cmux.HTTP1Fast())
	grpcL := tcpm.Match(cmux.Any())

	db := initDB(envCfg)
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Severef("failed to close db, err: %v", err)
		}
	}(db)
	s.invoke(envCfg, db)

	errGrp, ctx := errgroup.WithContext(ctx)
	errGrp.Go(func() error {
		s.runGrpcServer(ctx, grpcL)
		return nil
	})
	errGrp.Go(func() error {
		go s.runHttpServer(ctx, httpL, envCfg)
		return nil
	})
	if envCfg.Migration.Auto {
		if err := s.migrateUseCase.Up(ctx); err != nil {
			logger.Severef("failed to migrate db, err: %v", err)
		}
	}
	if envCfg.Seed.Auto {
		if err := s.seedUseCase.Seed(ctx); err != nil {
			logger.Severef("failed to seed, err: %v", err)
		}
	}
	go func() {
		if err := tcpm.Serve(); !strings.Contains(err.Error(), "use of closed") {
			logger.Severef("failed to serve: %v", err)
		}
	}()
	defer tcpm.Close()
	return errGrp.Wait()
}

func (s *server) invoke(envCfg config.EnvConfig, db *sql.DB) {
	var grpcSvr *grpc.Server
	var httpMux *runtime.ServeMux
	var migrateUseCase develop.MigrateUseCase
	var seedUseCase develop.SeedUseCase
	c := di.NewDI(envCfg, db)
	if err := c.Invoke(func(
		g *grpc.Server,
		m *runtime.ServeMux,
		muc develop.MigrateUseCase,
		suc develop.SeedUseCase,
	) {
		grpcSvr = g
		httpMux = m
		migrateUseCase = muc
		seedUseCase = suc
	}); err != nil {
		logger.Severef("di invoke err: %v", err)
	}
	s.grpcServer = grpcSvr
	s.httpServeMux = httpMux
	s.migrateUseCase = migrateUseCase
	s.seedUseCase = seedUseCase
}

func (s *server) runGrpcServer(ctx context.Context, lis net.Listener) {
	_ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("starting grpc server...")
		if err := s.grpcServer.Serve(lis); err != nil {
			logger.Severef("failed to serve: %v", err)
		}
	}()

	// Listen for the interrupt signal.
	<-_ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	logger.Info("grpc server shutting down gracefully")
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		// close listeners to stop accepting new connections,
		// will block on any existing transports
		s.grpcServer.GracefulStop()
	}()
	select {
	case <-ch:
		logger.Infof("grpc server graceful stopped")
	case <-time.After(10 * time.Second):
		// took too long, manually close open transports
		// e.g. watch streams
		logger.Infof("grpc server graceful stop timeout, force stop!!")
		s.grpcServer.Stop()
		<-ch
	}
	logger.Info("grpc server successfully stopped")
}

func (s *server) runHttpServer(ctx context.Context, lis net.Listener, envCfg config.EnvConfig) {
	_ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Create SPA handler for serving the frontend
	spaHandler := httphandler.NewSPAHandler()

	// Create a multiplexer that routes API requests to gRPC-Gateway and everything else to SPA
	mux := http.NewServeMux()

	// Route API requests (starting with /api/) to gRPC-Gateway
	mux.Handle("/api/", s.httpServeMux)

	// Route all other requests to SPA handler
	mux.Handle("/", spaHandler)

	// Apply CORS middleware
	withCors := cors.New(cors.Options{
		AllowedOrigins:   envCfg.CORS.AllowOrigins,
		AllowedMethods:   envCfg.CORS.AllowMethods,
		AllowedHeaders:   envCfg.CORS.AllowHeaders,
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: envCfg.CORS.AllowCredentials,
		MaxAge:           envCfg.CORS.MaxAge,
	}).Handler(mux)

	svr := &http.Server{
		Handler: withCors,
	}
	go func() {
		logger.Info("starting http server with SPA support...")
		if err := svr.Serve(lis); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Severef("http serve error: %v", err)
		}
	}()

	<-_ctx.Done()
	stop()
	logger.Info("http server shutting down gracefully")

	_ctxTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := svr.Shutdown(_ctxTimeout); err != nil {
		logger.Severef("http server forced to shutdown: %v", err)
	}

	logger.Info("http server successfully stopped")
}

func initDB(envCfg config.EnvConfig) *sql.DB {
	db, err := mysql.NewConn(envCfg.MySQL)
	if err != nil {
		logger.Severef("failed connect db, err: %v", err)
	}
	return db
}
