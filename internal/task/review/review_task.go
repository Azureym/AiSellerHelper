package review

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"PulseCheck/internal/tools"
	"PulseCheck/internal/xhsreq"
)

type ReviewReplyData struct {
	ReviewIds     []string
	ReviewContent string
	ReplyContent  string
}

type OrderIdReviewProvider struct {
	searchParam   *xhsreq.ReviewSearchParam
	reviewManager *xhsreq.ReviewManager
	xhsReviewChat *xhsreq.XHSReviewChat
}

func NewReviewProvider(searchParam *xhsreq.ReviewSearchParam, reviewManager *xhsreq.ReviewManager, xhsReviewChat *xhsreq.XHSReviewChat) *OrderIdReviewProvider {
	return &OrderIdReviewProvider{searchParam: searchParam, reviewManager: reviewManager, xhsReviewChat: xhsReviewChat}
}

func NewOrderIdReviewProvider(ctx context.Context, orderId string, reviewManager *xhsreq.ReviewManager, xhsReviewChat *xhsreq.XHSReviewChat) *OrderIdReviewProvider {
	param := &xhsreq.ReviewSearchParam{
		OrderID: orderId,
	}
	return &OrderIdReviewProvider{searchParam: param, reviewManager: reviewManager, xhsReviewChat: xhsReviewChat}
}

func (this *OrderIdReviewProvider) Provide(ctx context.Context) (<-chan []*ReviewReplyData, <-chan error) {
	replyDataChan := make(chan []*ReviewReplyData, 10)
	errChan := make(chan error)
	go func() {
		defer func() {
			close(replyDataChan)
			close(errChan)
		}()
		reviewSearchParam := this.searchParam
		tools.LogFromContext(ctx, "--正在获取获取小红书商品评价信息--")
		tools.LogFromContext(ctx, "searchParam: %#v", *this.searchParam)
		reviews, err := this.reviewManager.GetReviews(ctx, reviewSearchParam)
		if err != nil {
			errChan <- err
			return
		}
		tools.LogFromContext(ctx, "\n--获取成功--")
		for _, review := range reviews {
			tools.LogFromContext(ctx, "评论ID:%s", review.Id)
			tools.LogFromContext(ctx, "评论信息:%s", review.Content)
			tools.LogFromContext(ctx, "sku信息:%#v", *review.SkuInfo)
			tools.LogFromContext(ctx, "评分信息:%#v", *review.Score)
		}

		tools.LogFromContext(ctx, "\n--正在获取回复生成内容--")
		reviewReplyDataList := make([]*ReviewReplyData, 0, len(reviews))
		for _, review := range reviews {
			param := &xhsreq.XHSReviewChatParam{
				ItemId:        review.SkuInfo.ItemID,
				ItemInfo:      review.SkuInfo.SkuName,
				ReviewContent: review.Content,
			}

			answer, err := this.xhsReviewChat.Interact(ctx, param)
			if err != nil {
				errChan <- err
				return
			}
			reviewReplyData := &ReviewReplyData{
				ReviewIds:     []string{review.Id},
				ReviewContent: review.Content,
				ReplyContent:  answer,
			}
			reviewReplyDataList = append(reviewReplyDataList, reviewReplyData)
			tools.LogFromContext(ctx, "\n--获取成功--")
			tools.LogFromContext(ctx, "reviewId:%+v", reviewReplyData.ReviewIds)
			tools.LogFromContext(ctx, "reviewContent:%s", reviewReplyData.ReviewContent)
			tools.LogFromContext(ctx, "skuInfo:%#v", *review.SkuInfo)
			tools.LogFromContext(ctx, "score:%#v", *review.Score)
			tools.LogFromContext(ctx, "answer:%s", reviewReplyData.ReplyContent)
		}
		replyDataChan <- reviewReplyDataList
	}()
	return replyDataChan, errChan
}

// ----------------------------------
type ReviewReplyHandler struct {
	reviewReply *xhsreq.ReviewReply
}

func NewReviewReplyHandler(ctx context.Context, reviewReply *xhsreq.ReviewReply) *ReviewReplyHandler {
	return &ReviewReplyHandler{reviewReply: reviewReply}
}

func (this *ReviewReplyHandler) Execute(ctx context.Context, data []*ReviewReplyData) error {
	if len(data) == 0 {
		return nil
	}
	tools.LogFromContext(ctx, "\n--正在发起小红书回复--")

	var result error
	for _, reviewReply := range data {
		param := &xhsreq.ReviewReplyParam{
			ReviewIds:    reviewReply.ReviewIds,
			ReplyContent: reviewReply.ReplyContent,
		}
		tools.LogFromContext(ctx, "reviewIds:%v", reviewReply.ReviewIds)
		tools.LogFromContext(ctx, "reply:%s", reviewReply.ReplyContent)

		err := this.reviewReply.Reply(ctx, param)
		if err != nil {
			result = multierror.Append(result, errors.WithMessagef(err, "request param:%#v", *param))
			tools.LogFromContext(ctx, "--回复失败--")
			tools.LogFromContext(ctx, "error:%#v", err)
		}
		tools.LogFromContext(ctx, "\n--回复成功--")
	}
	return result
}
