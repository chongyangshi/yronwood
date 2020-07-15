package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/monzo/slog"

	"github.com/monzo/typhon"

	"github.com/chongyangshi/yronwood/config"
	"github.com/chongyangshi/yronwood/endpoints"
)

func main() {
	initContext := context.Background()
	svc := endpoints.Service()
	srv, err := typhon.Listen(svc, config.ConfigListenAddr)
	if err != nil {
		panic(err)
	}
	slog.Info(initContext, "Yronwood listening on %v", srv.Listener().Addr())

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	<-done
	slog.Info(initContext, "Yronwood shutting down")
	c, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Stop(c)
}
