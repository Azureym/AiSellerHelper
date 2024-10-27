package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
)

var (
	srv atomic.Pointer[http.Server]
)

type HttpFunc func(http.ResponseWriter, *http.Request)

func StartHttpServer(ctx context.Context, port int, handlers map[string]HttpFunc) error {
	mux := http.NewServeMux()
	for url, httpFunc := range handlers {
		mux.HandleFunc(url, httpFunc)
	}
	localSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}
	srv.Store(localSrv)
	go func() {
		log.Printf("http server listening on %s\n", localSrv.Addr)
		if err := localSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s error occured.\n", err)
		}
	}()

	return nil
}

func StopHttpServer(ctx context.Context) error {
	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	server := srv.Load()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")
	return nil
}

var (
	cr atomic.Pointer[cron.Cron]
)

type CronFunc func()

func StartCronServer(ctx context.Context, handlers map[string]CronFunc) error {
	c := cron.New(cron.WithLogger(
		cron.VerbosePrintfLogger(log.New(os.Stdout, "[CRON] ", log.LstdFlags)),
	))
	for spec, handle := range handlers {
		c.AddFunc(spec, handle)
	}
	c.Start()
	cr.Store(c)
	return nil
}

func StopCronServer(ctx context.Context, timeout time.Duration) error {
	stopCtx := cr.Load().Stop()
	select {
	case <-stopCtx.Done():
		log.Println("cron job has been stopped")
	case <-time.After(timeout):
		log.Println("stopping cron job due to context timeout.")
	}
	return nil
}
