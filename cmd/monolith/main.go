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
	"github.com/gorilla/mux"
	"github.com/oklog/run"
	"github.com/peterbourgon/gattaca/pkg/auth"
	"github.com/peterbourgon/gattaca/pkg/dna"
	"github.com/peterbourgon/usage"
	"github.com/pkg/errors"
)

func main() {
	fs := flag.NewFlagSet("monolith", flag.ExitOnError)
	var (
		apiAddr = fs.String("api", "127.0.0.1:8080", "HTTP API listen address")
		authURN = fs.String("auth-urn", "file:auth.db", "URN for auth DB")
		dnaURN  = fs.String("dna-urn", "file:dna.db", "URN for DNA DB")
	)
	fs.Usage = usage.For(fs, "monolith [flags]")
	fs.Parse(os.Args[1:])

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	var authsvc auth.Service
	{
		authrepo, err := auth.NewSQLiteRepository(*authURN)
		if err != nil {
			logger.Log("during", "auth.NewSQLiteRepository", "err", err)
			os.Exit(1)
		}
		authsvc = auth.NewDefaultService(authrepo)
	}

	var authserver http.Handler
	{
		authserver = auth.NewHTTPServer(authsvc)
	}

	var dnasvc dna.Service
	{
		dnarepo, err := dna.NewSQLiteRepository(*dnaURN)
		if err != nil {
			logger.Log("during", "dna.NewSQLiteRepository", "err", err)
			os.Exit(1)
		}
		dnasvc = dna.NewDefaultService(dnarepo, authsvc) // don't need a client
	}

	var dnaserver http.Handler
	{
		dnaserver = dna.NewHTTPServer(dnasvc)
	}

	var api http.Handler
	{
		r := mux.NewRouter()
		r.PathPrefix("/auth/").Handler(http.StripPrefix("/auth", authserver))
		r.PathPrefix("/dna/").Handler(http.StripPrefix("/dna", dnaserver))
		api = r
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
