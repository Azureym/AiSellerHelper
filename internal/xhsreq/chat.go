package xhsreq

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"

	"PulseCheck/internal/tools"
)

var (
	// initialize item type - item id mapping
	itemIDTypeMapping = map[string]string{
		"6564c049474aad0001c7641a": "裤夹|通过裤夹可以将裤子挂起来收纳到衣柜中,简单方便,整齐,省空间",
		"660ced34f46f6600013ee667": "垃圾架|可以适配各种类型的外卖袋和购物袋颜值超高,清洗方便",
		"669b9f0c0916240001ff6ac4": "吊带衣架|一个顶20个的吊带衣架,挂的多,省空间,拿取方便",
		"65e0718b3f330b0001d94c29": "缩脖子衣架|衣架缩脖子设计,省空间,垂直空间节约8厘米",
		"66ad1c39274e530001234bf9": "收纳线槽|桌底收纳电线整齐,方便调整",
	}
)

type XHSReviewChatParam struct {
	ItemId        string `json:"item_id" validate:"required"`
	ItemInfo      string `json:"item_info" validate:"required"`
	ReviewContent string `json:"review_content" validate:"required"`
}

type XHSReviewChat struct {
	httpClient *http.Client
}

const (
	SkuInfoPath    Path = "sku_info"
	ItemTypePath   Path = "item_type"
	ItemIntroPath  Path = "item_introduction"
	ReviewInfoPath Path = "review_info"
	TextPath       Path = "text"

	QueryPath        Path = "query"
	ResponseModePath Path = "response_mode"
	UserPath         Path = "user"
	InputPath        Path = "inputs"
	ConversationPath Path = "conversation_id"
)

func (this *XHSReviewChat) newRequestBody(ctx context.Context, param *XHSReviewChatParam) ([]byte, error) {
	queryJsonData := []byte("{}")
	tools.LogFromContext(ctx, "\n--正在转化商品信息--")
	itemTypeInfo, ok := itemIDTypeMapping[param.ItemId]
	if !ok {
		return nil, errors.Errorf("could find itemtype by itemid. SearchParam:%#v", param)
	}
	splitItemTypeInfo := strings.Split(itemTypeInfo, "|")
	tools.LogFromContext(ctx, "itemId:%s -> itemType:%s", param.ItemId, splitItemTypeInfo[0])
	tools.LogFromContext(ctx, "text:%s", splitItemTypeInfo[1])

	queryJsonData, _ = sjson.SetBytes(queryJsonData, SkuInfoPath.Join(ItemTypePath).String(), splitItemTypeInfo[0])
	queryJsonData, _ = sjson.SetBytes(queryJsonData, SkuInfoPath.Join(ItemIntroPath).String(), splitItemTypeInfo[1])
	queryJsonData, _ = sjson.SetBytes(queryJsonData, ReviewInfoPath.Join(TextPath).String(), param.ReviewContent)

	difyData := []byte("{}")
	difyData, _ = sjson.SetBytes(difyData, QueryPath.String(), queryJsonData)
	difyData, _ = sjson.SetBytes(difyData, ResponseModePath.String(), "blocking")
	difyData, _ = sjson.SetBytes(difyData, UserPath.String(), "abc-123")
	difyData, _ = sjson.SetBytes(difyData, ConversationPath.String(), "")
	difyData, _ = sjson.SetRawBytes(difyData, InputPath.String(), []byte("{}"))

	return difyData, nil
}

func (this *XHSReviewChat) validate(ctx context.Context, param *XHSReviewChatParam) error {
	validate := validator.New()
	err := validate.Struct(param)
	return errors.WithMessagef(err, "param:%#v", param)
}

const (
	difyKey = "Bearer app-SpZD4UXXSR5AQsyq5FF5ohZ9"
)

func (this *XHSReviewChat) Interact(ctx context.Context, param *XHSReviewChatParam) (string, error) {
	err := this.validate(ctx, param)
	if nil != err {
		return "", err
	}
	body, err := this.newRequestBody(ctx, param)
	bodyStr := string(body)
	slog.Info(spew.Sprintf("new request body:%s with param:%#v", bodyStr, param))
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.dify.ai/v1/chat-messages", bytes.NewBuffer(body))
	if nil != err {
		return "", errors.WithMessagef(err, "new request error. body:%s", bodyStr)
	}
	request.Header.Set("Authorization", difyKey)
	request.Header.Set("Content-Type", "application/json")

	response, err := this.httpClient.Do(request)
	if nil != err {
		return "", errors.WithMessagef(err, "request failed with param:%#v requestBody:%s", param, bodyStr)
	}
	statusCode := response.StatusCode
	if statusCode != http.StatusOK {
		slog.Error("get dify response error.",
			slog.Int("statusCode", statusCode),
			slog.String("body", bodyStr),
		)
		return "", errors.Errorf("get dify response error. difyURL:%s statusCode:%d requestBody:%s",
			DifyDomain, statusCode, bodyStr)
	}
	respBody := response.Body
	defer func() {
		err := respBody.Close()
		if err != nil {
			slog.Error("close response body err.", tools.ErrAttr(err))
		}
	}()
	b, err := io.ReadAll(respBody)
	if err != nil {
		return "", errors.WithMessagef(err, "read response body fail with param:%#v body:%#v", param, bodyStr)
	}

	return this.withdrawAnswer(ctx, b)
}

const (
	AnswerPath Path = "answer"
)

func (this *XHSReviewChat) withdrawAnswer(ctx context.Context, response []byte) (string, error) {
	jsonData := gjson.ParseBytes(response)
	result := jsonData.Get(AnswerPath.String())
	if !result.Exists() {
		return "", errors.Errorf("answer not exists. response:%s", jsonData.String())
	}
	return result.String(), nil
}

func NewXHSReviewChat(ctx context.Context, httpClient *http.Client) *XHSReviewChat {
	return &XHSReviewChat{httpClient: httpClient}
}

func NewXHSReviewChatWithHTTP(ctx context.Context) *XHSReviewChat {
	return &XHSReviewChat{httpClient: tools.NewHttpsClient(DifyDomain, tools.WithTimeout(2*time.Minute))}
}
