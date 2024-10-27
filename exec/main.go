package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"

	"PulseCheck/internal/config"
	"PulseCheck/internal/tools"
)

const (
	RandomMinutesRange = 10
	ProgramTimeout     = 5 * time.Minute
	HttpServerPort     = 1903
)

func main() {
	config.InitLogger()
	defer func() {
		err := config.CloseLogger()
		if nil != err {
			log.Fatal("close log error")
		}
	}()
	config.StartPprof(nil)
	executeAndWaitExit([]os.Signal{syscall.SIGTERM, syscall.SIGKILL}, func(ctx context.Context) error {
		// start cron server
		if err := StartCronServer(ctx, map[string]CronFunc{
			// execute every day at noon
			"0 0 * * ?": func() {
				err := ReplyForLatestReview(ctx)
				if nil != err {
					slog.Error("check run error", tools.ErrAttr(err))
				}
			},
		}); err != nil {
			log.Fatalf("start cron server. error: %v", err)
		}
		slog.Info("cron server started")
		// start http server
		if err := StartHttpServer(ctx, HttpServerPort, map[string]HttpFunc{
			"/replywithorderid": func(writer http.ResponseWriter, request *http.Request) {
				ctx1 := tools.AppendXHSToken(ctx, authorization)
				ctx1 = tools.AppendWriter(ctx1, writer)
				orderID := request.FormValue("orderid")
				if len(orderID) == 0 {
					tools.LogFromContext(ctx, "orderid param not set properly")
					return
				}
				ReplyWithOrderID(ctx1, orderID)
			},
		}); err != nil {
			log.Fatalf("start http server. port:%d error: %v", HttpServerPort, err)
		}
		slog.Info("http server started")

		<-ctx.Done()
		slog.Info("context canceled, shutting down server.", slog.String("cause", context.Cause(ctx).Error()))

		slog.Info("stopping cron job.")
		// ----------- stop cron
		if err := StopCronServer(ctx, ProgramTimeout); err != nil {
			slog.Error("stop cron server. error: %v", tools.ErrAttr(err))
		}
		slog.Info("cron server has been stopped")
		// ---------- stop http server
		slog.Info("stopping http server.")
		if err := StopHttpServer(ctx); err != nil {
			slog.Error("stop http server.", slog.Int("port", HttpServerPort), tools.ErrAttr(err))
		}
		slog.Info("http server has been stopped")
		return nil
	})
}

func executeAndWaitExit(signals []os.Signal, run func(ctx context.Context) error) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, signals...)
	ctx, cancelCausedBy := context.WithCancelCause(context.Background())
	done := make(chan struct{})
	go func() {
		err := run(ctx)
		if nil != err {
			log.Fatalf("error:%+v program exit.\n", err)
		}
		done <- struct{}{}
	}()
LOOP:
	for {
		select {
		case signalvar := <-c:
			slog.Info("program exit", slog.String("exit code", signalvar.String()))
			slog.Info("waiting for other goroutines to exit...")
			cancelCausedBy(errors.Errorf("cancelled from exit code:%s", signalvar.String()))
		case <-done:
			slog.Info("other goroutines exit successfully.")
			slog.Info("============== program exit ===================")
			break LOOP
		}
	}
	os.Exit(0)
}
