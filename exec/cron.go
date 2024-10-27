package main

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/pkg/errors"

	"PulseCheck/internal/task"
	"PulseCheck/internal/task/alarm"
)

func cronCheckLogistics(ctx context.Context) error {
	slog.Info("cron task has started...")
	defer slog.Info("cron task has finished.")
	xhsDeliveryTask := task.NewTask[*alarm.XHSDeliveryStatistics](
		alarm.NewXiaohongshuStatisticsProvider(),
		alarm.NewXHSStatisticsHandler(),
		alarm.NewXHSStatisticsDataFilter(),
	)
	err := xhsDeliveryTask.Execute(ctx)
	return errors.WithMessagef(err, "xiaohongshu delivery task error.")
}

func checkLogistics(ctx context.Context) error {
	slog.Info("test task has started...")
	defer slog.Info("test task has finished.")
	xhsDeliveryTask := task.NewTask[*alarm.XHSDeliveryStatistics](
		alarm.NewXiaohongshuStatisticsProvider(),
		alarm.NewXHSStatisticsHandler(),
	)
	err := xhsDeliveryTask.Execute(ctx)
	return errors.WithMessagef(err, "xiaohongshu delivery task error.")
}
func withRandDelay(delayMinutesRange int, do func(ctx context.Context) error) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		intn := r.Intn(delayMinutesRange)
		timer := time.NewTimer(time.Duration(intn+1) * time.Minute)
		slog.Info("the cron task will be delayed.", slog.Int("periodMinutes", intn))
		<-timer.C
		return do(ctx)
	}
}
