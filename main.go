package main

import (
	"fmt"
	"net/http"

	"github.com/btschwartz12/autodeploy/server"
	flags "github.com/jessevdk/go-flags"
	"go.uber.org/zap"
)

type arguments struct {
	DevLogging bool   `short:"d" long:"dev-logging" description:"Enable development logging"`
	Port       int    `short:"p" long:"port" description:"Port to listen on" default:"8000"`
	ConfigPath string `short:"c" long:"config" description:"Path to config file" default:"config.yaml" env:"AUTODEPLOY_CONFIG_PATH"`
}

var args arguments

func main() {
	_, err := flags.Parse(&args)
	if err != nil {
		panic(fmt.Errorf("error parsing flags: %w", err))
	}

	var l *zap.Logger
	if args.DevLogging {
		l, _ = zap.NewDevelopment()
	} else {
		l, _ = zap.NewProduction()
	}
	logger := l.Sugar()

	if args.Port == 0 {
		logger.Fatalw("port must be set")
	}

	s, err := server.NewServer(logger, args.ConfigPath)
	if err != nil {
		logger.Fatalw("failed to create server", "error", err)
	}

	errChan := make(chan error)
	go func() {
		logger.Infow("Starting server", "port", args.Port)
		errChan <- http.ListenAndServe(fmt.Sprintf(":%d", args.Port), s.GetRouter())
	}()
	err = <-errChan
	logger.Fatalw("http server failed", "error", err)
}
