package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	v1 "github.com/mt-sre/addon-service/internal/service/v1"
	addonsv1 "github.com/mt-sre/addon-service/internal/service/v1/addons"
	"github.com/mt-sre/addon-service/internal/service/v1/version"
)

type serverConfig struct {
	// grpcPort is the port on which the gRPC server will listen
	grpcPort string

	// httpPort is the port on which the HTTP server will listen
	httpPort string
}

// runServer runs the HTTP and gRPC server
func runServer() error {
	ctx := context.Background()

	var cfg serverConfig
	flag.StringVar(&cfg.grpcPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&cfg.grpcPort, "http-port", "", "gRPC port to bind")
	flag.Parse()

	if len(cfg.grpcPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", cfg.grpcPort)
	}

	if len(cfg.grpcPort) == 0 {
		return fmt.Errorf("invalid TCP port for HTTP gateway: '%s'", cfg.grpcPort)
	}

	var (
		addonv1API   = addonsv1.NewAddonService()
		versionv1API = version.NewVersionService()
	)

	errChan := make(chan error)

	// run HTTP server
	go func(errChan chan error) {
		err := v1.RunHTTPServer(
			ctx,
			cfg.grpcPort,
			cfg.httpPort,
		)
		errChan <- err
	}(errChan)

	if err := <-errChan; err != nil {
		return err
	}

	return v1.RunGRPCServer(
		ctx,
		cfg.grpcPort,
		addonv1API,
		versionv1API,
	)
}

func main() {
	if err := runServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
