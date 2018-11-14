package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/oklog/run"
	"github.com/peterbourgon/gattaca/pkg/dna"
	"github.com/peterbourgon/usage"
	"github.com/pkg/errors"
)

func main() {
	fs := flag.NewFlagSet("dnasvc", flag.ExitOnError)
	var (
		apiAddr     = fs.String("api", "127.0.0.1:8082", "HTTP API listen address")
		urn         = fs.String("urn", "dna.db", "URN for DNA DB")
		authsvcAddr = fs.String("authsvc", "http://127.0.0.1:8081", "HTTP endpoint for authsvc")
	)
	fs.Usage = usage.For(fs, "dnasvc [flags]")
	fs.Parse(os.Args[1:])

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	var dnarepo dna.Repository
	{
		var err error
		dnarepo, err = dna.NewSQLiteRepository(*urn)
		if err != nil {
			logger.Log("during", "dna.NewSQLiteRepository", "err", err)
			os.Exit(1)
		}
	}

	var validator dna.Validator
	{
		validator = authClient(*authsvcAddr)
	}

	var dnasvc dna.Service
	{
		dnasvc = dna.NewDefaultService(dnarepo, validator)
	}

	var api http.Handler
	{
		api = dna.NewHTTPServer(dnasvc)
	}

	var g run.Group
	{
		server := &http.Server{
			Addr:    *apiAddr,
			Handler: api,
		}
		g.Add(func() error {
			logger.Log("component", "API", "addr", *apiAddr)
			return server.ListenAndServe()
		}, func(error) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			server.Shutdown(ctx)
		})
	}
	{
		ctx, cancel := context.WithCancel(context.Background())
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case sig := <-c:
				return errors.Errorf("received signal %s", sig)
			}
		}, func(error) {
			cancel()
		})
	}
	logger.Log("exit", g.Run())
}
