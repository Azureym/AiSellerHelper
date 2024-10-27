package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/pkg/errors"

	"PulseCheck/internal/task"
	"PulseCheck/internal/task/review"
	"PulseCheck/internal/tools"
	"PulseCheck/internal/xhsreq"
)

func ReplyWithOrderID(ctx context.Context, orderID string) error {
	slog.Info("cron task has started...")
	defer slog.Info("cron task has finished.")
	xhsHttpsClient := tools.NewHttpsClient(xhsreq.XiaohongshuDomain)

	reviewManager := xhsreq.NewReviewManager(ctx, xhsHttpsClient)
	reviewChat := xhsreq.NewXHSReviewChatWithHTTP(ctx)
	reviewReply := xhsreq.NewReviewReply(ctx, xhsHttpsClient)

	xhsReviewReplyTask := task.NewTask[[]*review.ReviewReplyData](
		review.NewOrderIdReviewProvider(ctx, orderID, reviewManager, reviewChat),
		review.NewReviewReplyHandler(ctx, reviewReply),
	)
	err := xhsReviewReplyTask.Execute(ctx)
	return errors.WithMessagef(err, "xiaohongshu delivery task error.")
}

func ReplyForLatestReview(ctx context.Context) error {
	slog.Info("cron task has started...")
	defer slog.Info("cron task has finished.")
	xhsHttpsClient := tools.NewHttpsClient(xhsreq.XiaohongshuDomain)

	reviewManager := xhsreq.NewReviewManager(ctx, xhsHttpsClient)
	reviewChat := xhsreq.NewXHSReviewChatWithHTTP(ctx)
	reviewReply := xhsreq.NewReviewReply(ctx, xhsHttpsClient)

	after := time.Now()
	start := after.Add(-24 * time.Hour)
	param := &xhsreq.ReviewSearchParam{
		ContentTypeList:       []int{2},
		ReviewReplyStatusList: []int{2},
		StartTime:             &start,
		EndTime:               &after,
		PageSize:              20,
	}
	xhsReviewReplyTask := task.NewTask[[]*review.ReviewReplyData](
		review.NewReviewProvider(param, reviewManager, reviewChat),
		review.NewReviewReplyHandler(ctx, reviewReply),
	)
	err := xhsReviewReplyTask.Execute(ctx)
	return errors.WithMessagef(err, "xiaohongshu delivery task error.")
}
