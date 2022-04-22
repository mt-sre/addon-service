package v1

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/mt-sre/addon-service/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// RunGRPCServer registers services and runs a gRPC server
func RunGRPCServer(ctx context.Context, port string,
	addonsAPI v1.AddonServiceServer,
	versionsAPI v1.VersionServiceServer) error {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		return err
	}

	// gRPC server options
	opts := []grpc.ServerOption{}

	// register services
	server := grpc.NewServer(opts...)
	v1.RegisterAddonServiceServer(server, addonsAPI)
	v1.RegisterVersionServiceServer(server, versionsAPI)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			log.Println("shutting down gRPC server...")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	log.Printf("starting gRPC server on port %s\n", port)
	return server.Serve(listen)
}

// RunHTTPServer runs a HTTP server
func RunHTTPServer(ctx context.Context, grpcPort, httpPort string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	if err := v1.RegisterAddonServiceHandlerFromEndpoint(ctx, mux,
		fmt.Sprintf("localhost:%s", grpcPort), opts); err != nil {
		log.Fatalf("failed to start HTTP gateway: %v", err)
	}

	if err := v1.RegisterVersionServiceHandlerFromEndpoint(ctx, mux,
		fmt.Sprintf("localhost:%s", grpcPort), opts); err != nil {
		log.Fatalf("failed to start HTTP gateway: %v", err)
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", httpPort),
	}

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
		}

		_, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_ = srv.Shutdown(ctx)
	}()

	log.Printf("starting HTTP server on port %s\n", httpPort)
	return srv.ListenAndServe()
}
