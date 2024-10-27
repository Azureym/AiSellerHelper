package alarm

import (
	"bytes"
	"context"
	"crypto/tls"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"gopkg.in/gomail.v2"

	"PulseCheck/internal/task"
	"PulseCheck/internal/tools"
)

type XHSDeliveryStatistics struct {
	Total                        int `json:"total"`
	CollectTimeoutCount          int `json:"collect_timeout_count"`
	CollectWarnCount             int `json:"collect_warn_count"`
	ReturnRejectCount            int `json:"return_reject_count"`
	LogisticsStandstillCount     int `json:"logistics_standstill_count"`
	CollectTransportWarnCount    int `json:"collect_transport_warn_count"`
	CollectTransportTimeoutCount int `json:"collect_transport_timeout_count"`
	DeliverySignTimeoutCount     int `json:"delivery_sign_timeout_count"`
	LogisticsRouteTimeoutCount   int `json:"logistics_route_timeout"`
}

func (this *XHSDeliveryStatistics) ValidTotalData() int {
	return this.Total - this.ReturnRejectCount - this.LogisticsRouteTimeoutCount - this.DeliverySignTimeoutCount - this.LogisticsStandstillCount
}

func (x *XHSDeliveryStatistics) String() string {
	marshal, e := json.Marshal(x)
	if e != nil {
		slog.Error("Failed to marshal XiaohongshuDeliveryStatistics", tools.ErrAttr(e))
		return "[XiaohongshuDeliveryStatistics error]"
	}
	return string(marshal)
}

type XiaohongshuStatisticsProvider struct {
	httpClient *http.Client
}

const (
	xiaohongshuDomain = "ark.xiaohongshu.com"

	httpClientMaxIdleConns        = 20
	httpClientMaxConnsPerHost     = 20
	httpClientMaxIdleConnsPerHost = 20
	httpClientTimeout             = 10 * time.Second
)

func NewXiaohongshuStatisticsProvider() *XiaohongshuStatisticsProvider {

	t := http.DefaultTransport.(*http.Transport).Clone()
	t.MaxIdleConns = httpClientMaxIdleConns
	t.MaxConnsPerHost = httpClientMaxConnsPerHost
	t.MaxIdleConnsPerHost = httpClientMaxIdleConnsPerHost

	t.TLSClientConfig = &tls.Config{
		RootCAs:    tools.SystemCertPool(),
		ServerName: xiaohongshuDomain,
	}

	return &XiaohongshuStatisticsProvider{
		httpClient: &http.Client{
			Timeout:   httpClientTimeout,
			Transport: t,
		},
	}
}

const (
	xiaohongshuLogisticsStatisticsURL = "https://ark.xiaohongshu.com/api/edith/logistics/abnormal/count"
)

/*
Sample:

Request:
curl 'https://ark.xiaohongshu.com/api/edith/logistics/abnormal/count' \
  -H 'accept: application/json, text/plain, * /*' \
-H 'accept-language: en' \
-H 'authorization: AT-68c5173874366708082130495h6xuhfr31cwfhep' \
-H 'cache-control: no-cache' \
-H 'content-type: application/json' \
-H 'cookie: a1=18f137ccfed9g2hoelwpjqyvnunsiagjjyjd5stox00000135039; webId=a23bf7a341ef68df96c1281831cbd021; gid=yYiyqWSffiVSyYiyqWSSivM9dfjkJxld1EUVVlCAvhI7If88M30kV2888yq28qj80JS4i8i4; customerClientId=688073957365739; x-user-id-zhaoshang.xiaohongshu.com=; ark_worker_plugin_uuid=886235df2bf9422ebe1ea9c73915763c; x-user-id-school.xiaohongshu.com=65d6ecd9e300000000000001; abRequestId=a23bf7a341ef68df96c1281831cbd021; access-token=; sso-type=customer; subsystem=ark; x-user-id-ark.xiaohongshu.com=5679401044760815fb659cc6; web_session=040069431fa14842cc90b0bc41344b8ee07f5d; unread={%22ub%22:%22666a5e8400000000060064ec%22%2C%22ue%22:%226672820f0000000006004d36%22%2C%22uc%22:55}; customer-sso-sid=68c517387436666513245744e7b66615a3a7ed7e; access-token-ark.xiaohongshu.com=customer.ark.AT-68c5173874366708082130495h6xuhfr31cwfhep; access-token-ark.beta.xiaohongshu.com=customer.ark.AT-68c5173874366708082130495h6xuhfr31cwfhep; webBuild=4.24.2; xsecappid=xhs-pc-web; websectiga=3633fe24d49c7dd0eb923edc8205740f10fdb18b25d424d2a2322c6196d2a4ad; acw_tc=6c0bbba5637146221a6d3eea6f0da928938da4f613e68a98f8dd2b9331faf66b; beaker.session.id=86b4dd76c9c3d900e9b118602f97f1b2d30950cbgAJ9cQEoWA4AAAByYS11c2VyLWlkLWFya3ECWBgAAAA2NWQ2ZWNkOWUzMDAwMDAwMDAwMDAwMDFxA1UIX2V4cGlyZXNxBGNkYXRldGltZQpkYXRldGltZQpxBVUKB+gIBQIgDwmwwIVScQZYCwAAAGFyay1saWFzLWlkcQdYGAAAADY1NWYyZjViM2VhZDRmMDAwMTVhZTZjMXEIWA4AAABfYWNjZXNzZWRfdGltZXEJR0HZopQ9oTOcVQNfaWRxClggAAAANGM5NThlNmVjZjNhNDRmM2I4MWE5YWQ5MDIwMmRmNGNxC1gRAAAAcmEtYXV0aC10b2tlbi1hcmtxDFhBAAAAMzYxNmM5YzVjOGU0NDJiMGFmMjU1YjM4YjJmMThhYzktMGM4M2JjMTM2ODdjNDhlMzk4YWVhOTJiNjRmMGFjMDhxDVgOAAAAX2NyZWF0aW9uX3RpbWVxDkdB2aFcoOzdL3Uu' \
-H 'dnt: 1' \
-H 'origin: https://ark.xiaohongshu.com' \
-H 'pragma: no-cache' \
-H 'priority: u=1, i' \
-H 'referer: https://ark.xiaohongshu.com/app-order/abnormal/order/logistics' \
-H 'sec-ch-ua: "Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"' \
-H 'sec-ch-ua-mobile: ?0' \
-H 'sec-ch-ua-platform: "macOS"' \
-H 'sec-fetch-dest: empty' \
-H 'sec-fetch-mode: cors' \
-H 'sec-fetch-site: same-origin' \
-H 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36' \
-H 'x-b3-traceid: eb0416aad8a8cf9f' \
-H 'x-s: sjsWZBOksgvb02qvs2dUO2sCO25bOl5WZjs+ZjdBslM3' \
-H 'x-t: 1720340912549' \
--data-raw '{"status":[100],"marked":0,"shipment_end_time":1720340911000,"shipment_start_time":1712564911000,"package_finished_flag":0}' | jq

-------------
Response:
{
  "code": 200,
  "success": true,
  "msg": "成功",
  "data": {
    "collect_transport_warn_count": 0,
    "collect_transport_timeout_count": 0,
    "logistics_standstill_count": 0,
    "logistics_route_timeout_count": 2,
    "total": 11,
    "collect_warn_count": 0,
    "collect_timeout_count": 0,
    "delivery_sign_timeout_count": 1,
    "return_reject_count": 8
  }
}
*/
func (this *XiaohongshuStatisticsProvider) Provide(ctx context.Context) (<-chan *XHSDeliveryStatistics, <-chan error) {
	ch := make(chan *XHSDeliveryStatistics)
	errChan := make(chan error)
	go func() {
		defer func() {
			close(ch)
			close(errChan)
		}()
		requestBody, err2 := this.generateRequestBody(ctx)
		if nil != err2 {
			errChan <- errors.WithMessagef(err2, "generate request body error.")
			return
		}
		slog.Info("request body", slog.String("request body", requestBody))
		requestBodyReader := bytes.NewReader([]byte(requestBody))
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, xiaohongshuLogisticsStatisticsURL, requestBodyReader)
		if nil != err {
			errChan <- errors.WithMessagef(err, "construct xiaohongshu statistics request error.")
			return
		}

		err = this.setHeader(ctx, request.Header)
		if nil != err {
			errChan <- errors.WithMessagef(err, "http set header error.")
			return
		}

		response, err := this.httpClient.Do(request)
		if nil != err {
			errChan <- errors.WithMessagef(err, "access xiaohongshu error. xiaohongshuURL:%s", xiaohongshuLogisticsStatisticsURL)
			return
		}

		statusCode := response.StatusCode
		if statusCode != http.StatusOK {
			slog.Error("get xiaohongshu response error.",
				slog.String("xiaohognshuURL", xiaohongshuLogisticsStatisticsURL),
				slog.Int("statusCode", statusCode),
				slog.String("body", requestBody),
			)
			errChan <- errors.Errorf("get xiaohongshu response error. xiaohongshuURL:%s statusCode:%d requestBody:%s",
				xiaohongshuLogisticsStatisticsURL, statusCode, requestBody)
			return
		}
		respBody := response.Body
		defer func() {
			err := respBody.Close()
			if err != nil {
				slog.Error("close response body err.", tools.ErrAttr(err))
			}
		}()
		respData, err1 := this.unmarshal(respBody)
		if nil != err1 {
			errChan <- errors.WithMessagef(err1, "unmarshal body data error.")
			return
		}
		slog.Info("get response data from xiaohongshu.", slog.String("statistics", cast.ToString(respData)))
		ch <- respData
	}()
	return ch, errChan
}

type XHSStatisticsDataFilter struct {
}

func NewXHSStatisticsDataFilter() *XHSStatisticsDataFilter {
	return &XHSStatisticsDataFilter{}
}

func (this *XHSStatisticsDataFilter) DoFilter(ctx context.Context, data **XHSDeliveryStatistics, chain task.FilterChain[*XHSDeliveryStatistics]) error {
	if (*data).ValidTotalData() > 0 {
		slog.Info("statistics data is valid.", slog.String("statistics", cast.ToString(*data)))
		return chain.Proceed(ctx, data)
	}
	slog.Info("statistics data is invalid.", slog.String("statistics", cast.ToString(*data)))
	return nil
}

const (
	CODE_SUCCESS = 200
)

/*
json sample:
{
    "code": 200,
    "success": true,
    "msg": "成功",
    "data": {
        "collect_timeout_count": 0,
        "collect_transport_timeout_count": 0,
        "logistics_standstill_count": 0,
        "delivery_sign_timeout_count": 0,
        "return_reject_count": 1,
        "total": 1,
        "collect_warn_count": 0,
        "collect_transport_warn_count": 0,
        "logistics_route_timeout_count": 0
    }
}
*/
func (this *XiaohongshuStatisticsProvider) unmarshal(reader io.Reader) (*XHSDeliveryStatistics, error) {
	b, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	jsonData := gjson.ParseBytes(b)
	codeResule := jsonData.Get("code")
	if !codeResule.Exists() || codeResule.Int() != CODE_SUCCESS {
		msg := jsonData.Get("msg").String()
		return nil, errors.Errorf("response error:[%+v], jsondata:%+v", msg, jsonData.String())
	}
	dataJson := jsonData.Get("data")
	if !dataJson.Exists() || !dataJson.IsObject() {
		return nil, errors.Errorf("response error cannot get data element:[%+v]", jsonData.String())

	}
	xiaohongshuData := &XHSDeliveryStatistics{}
	total := cast.ToInt(jsonData.Get("data.total").Int())
	xiaohongshuData.Total = total

	collectTimeoutCount := cast.ToInt(jsonData.Get("data.collect_timeout_count").Int())
	xiaohongshuData.CollectTimeoutCount = collectTimeoutCount

	collectTransportTimeoutCount := cast.ToInt(jsonData.Get("data.collect_transport_timeout_count").Int())
	xiaohongshuData.CollectTransportTimeoutCount = collectTransportTimeoutCount

	logisticsStandstillCount := cast.ToInt(jsonData.Get("data.logistics_standstill_count").Int())
	xiaohongshuData.LogisticsStandstillCount = logisticsStandstillCount

	deliverySignTimeoutCount := cast.ToInt(jsonData.Get("data.delivery_sign_timeout_count").Int())
	xiaohongshuData.DeliverySignTimeoutCount = deliverySignTimeoutCount

	returnRejectCount := cast.ToInt(jsonData.Get("data.return_reject_count").Int())
	xiaohongshuData.ReturnRejectCount = returnRejectCount

	collectWarncount := cast.ToInt(jsonData.Get("data.collect_warn_count").Int())
	xiaohongshuData.CollectWarnCount = collectWarncount

	collectTransportWarnCount := cast.ToInt(jsonData.Get("data.collect_transport_warn_count").Int())
	xiaohongshuData.CollectTransportWarnCount = collectTransportWarnCount

	logisticsRouteTimeoutCount := cast.ToInt(jsonData.Get("data.logistics_route_timeout_count").Int())
	xiaohongshuData.LogisticsRouteTimeoutCount = logisticsRouteTimeoutCount

	return xiaohongshuData, nil
}

const (
	cookieFormat = `a1=18f137ccfed9g2hoelwpjqyvnunsiagjjyjd5stox00000135039; webId=a23bf7a341ef68df96c1281831cbd021; gid=yYiyqWSffiVSyYiyqWSSivM9dfjkJxld1EUVVlCAvhI7If88M30kV2888yq28qj80JS4i8i4; customerClientId=688073957365739; x-user-id-zhaoshang.xiaohongshu.com=; ark_worker_plugin_uuid=886235df2bf9422ebe1ea9c73915763c; x-user-id-school.xiaohongshu.com=65d6ecd9e300000000000001; abRequestId=a23bf7a341ef68df96c1281831cbd021; access-token=; sso-type=customer; subsystem=ark; web_session=040069431fa14842cc90b0bc41344b8ee07f5d; webBuild=4.25.1; unread={%22ub%22:%22668bd8750000000025015bc8%22%2C%22ue%22:%22668f74db0000000025002eed%22%2C%22uc%22:33}; acw_tc=2521b57b58b2b5ba5149cbd95cbeb0ba54231d009c1bef6bb34afbd638e02b91; xsecappid=sellercustomer; websectiga=f3d8eaee8a8c63016320d94a1bd00562d516a5417bc43a032a80cbf70f07d5c0; sec_poison_id=ae4e27b9-601e-40f1-b112-4351b82c64f1; customer-sso-sid=68c5173991171136170488473a6fe5c96bf49c16; x-user-id-ark.xiaohongshu.com=5679401044760815fb659cc6; access-token-ark.xiaohongshu.com=customer.ark.{{ . }}; access-token-ark.beta.xiaohongshu.com=customer.ark.{{ . }}; beaker.session.id=1d657de3051687a200364d516edbdc967440edc8gAJ9cQAoWAsAAABhcmstbGlhcy1pZHEBWBgAAAA2NTVmMmY1YjNlYWQ0ZjAwMDE1YWU2YzFxAlgOAAAAcmEtdXNlci1pZC1hcmtxA1gYAAAANjVkNmVjZDllMzAwMDAwMDAwMDAwMDAxcQRYDgAAAF9jcmVhdGlvbl90aW1lcQVHQdmrvHS+JN1YEQAAAHJhLWF1dGgtdG9rZW4tYXJrcQZYQQAAADM2MTZjOWM1YzhlNDQyYjBhZjI1NWIzOGIyZjE4YWM5LTBjODNiYzEzNjg3YzQ4ZTM5OGFlYTkyYjY0ZjBhYzA4cQdYAwAAAF9pZHEIWCAAAAA4MzhiZWNjZjE4ZWY0YTQ5OTZjZTQ1MWUxZTJlZmQwN3EJWA4AAABfYWNjZXNzZWRfdGltZXEKR0HZq7x0viTddS4=`
)

var (
	cookie        string
	authorization string
)

// auth=***** ./script.sh
func init() {
	auth, exists := os.LookupEnv("auth")
	if !exists {
		log.Fatalf("can not get auth key.")
	}
	log.Println(spew.Sprintf("get auth from env. auth=%s", auth))
	authorization = auth
	cookie = getCookie(authorization)
}
func getCookie(v string) string {
	t, b := template.New("cookieTemplate"), new(strings.Builder)
	err := template.Must(t.Parse(cookieFormat)).Execute(b, v)
	if err != nil {
		log.Fatalf("could not parse cookie. %s. error:%+v", cookieFormat, err)
	}
	return b.String()
}

func (this *XiaohongshuStatisticsProvider) setHeader(ctx context.Context, header http.Header) error {
	header.Set("accept", "application/json, text/plain, */*")
	header.Set("accept-language", "en")
	header.Set("authorization", authorization)
	header.Set("content-type", "application/json")
	header.Set("cookie", cookie)
	header.Set("dnt", "1")
	header.Set("origin", "https://ark.xiaohongshu.com")
	header.Set("referer", "https://ark.xiaohongshu.com/app-order/abnormal/order/logistics")
	header.Set("sec-fetch-dest", "empty")
	header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")
	header.Set("x-b3-traceid", "eb0416aad8a8cf9f")
	header.Set("x-s", "sjsWZBOksgvb02qvs2dUO2sCO25bOl5WZjs+ZjdBslM3")
	header.Set("x-t", cast.ToString(time.Now().UnixMilli()))
	header.Set("sec-ch-ua", `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`)
	header.Set("sec-ch-ua-mobile", `?0`)
	header.Set("sec-ch-ua-platform", `"macOS"`)
	header.Set("sec-fetch-dest", `empty`)
	header.Set("sec-fetch-mode", `cors`)
	header.Set("sec-fetch-site", `same-origin`)
	return nil
}

const (
	period = 7 * 24 * time.Hour
)

func (this *XiaohongshuStatisticsProvider) generateRequestBody(ctx context.Context, ) (string, error) {
	now := time.Now()
	end := now.UnixMilli()
	start := now.Add(-period).UnixMilli()
	body :=
		fmt.Sprintf(`
{
    "status": [
        100
    ],
    "marked": 0,
    "shipment_start_time": %d,
    "shipment_end_time": %d,
    "package_finished_flag": 0
}
`, start, end)
	return body, nil
}

type XHSStatisticsEmailHandler struct {
	// YANGMU_TODO: 2024/10/4 -- need to move the configure of the email info to here
}

func NewXHSStatisticsHandler() *XHSStatisticsEmailHandler {
	return &XHSStatisticsEmailHandler{}
}

var (
	toMail        = []string{"317128388@qq.com", "Azureym@126.com"}
	smtpServer    = "smtp.126.com"
	smtpPort      = 465
	subject       = "【小文同学】小红书店铺物流异常提醒"
	fromMail      = "Azureym@126.com"
	fromPasswd    = "a1b2c3d4Abcd"
	logisticsPage = "https://ark.xiaohongshu.com/app-order/abnormal/order/logistics"
)

func init() {
	tools.InitialCerts()
}

func (this *XHSStatisticsEmailHandler) sendEmail(from string, to []string, subject string, body string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", fromMail)
	message.SetHeader("To", to...)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)
	dialer := gomail.NewDialer(smtpServer, smtpPort, fromMail, fromPasswd)
	dialer.TLSConfig = &tls.Config{
		RootCAs:    tools.SystemCertPool(),
		ServerName: smtpServer,
	}
	if err := dialer.DialAndSend(message); err != nil {
		slog.Error("email message sending error.", tools.ErrAttr(err))
		return err
	}
	return nil
}

func (x *XHSStatisticsEmailHandler) Execute(ctx context.Context, data *XHSDeliveryStatistics) error {
	slog.Info("start to send email.", slog.String("data", cast.ToString(data)))
	body, err := x.createMsgBody(data)
	if nil != err {
		return errors.WithMessagef(err, "original data:%s", cast.ToString(data))
	}
	err1 := x.sendEmail(fromMail, toMail, subject, body)
	if nil != err1 {
		return errors.WithMessagef(err1, "send email error. original data:%s", cast.ToString(data))
	}
	return nil
}

//go:embed email_content.html
var templateContent string
var bodyTemplate *template.Template

func init() {
	bodyTemplate = template.Must(template.New("email logistics statistics template").Parse(templateContent))
}

func (this *XHSStatisticsEmailHandler) createMsgBody(statistics *XHSDeliveryStatistics) (string, error) {
	render := &XiaohongshuDeliveryStatisticsRender{
		Total:                        statistics.ValidTotalData(),
		CollectTimeoutCount:          statistics.CollectTimeoutCount,
		CollectWarnCount:             statistics.CollectWarnCount,
		CollectTransportWarnCount:    statistics.CollectTransportWarnCount,
		CollectTransportTimeoutCount: statistics.CollectTransportTimeoutCount,
	}
	var buf bytes.Buffer
	if err := bodyTemplate.Execute(&buf, map[string]any{
		"logisticsPageURL": logisticsPage,
		"data":             render,
	}); err != nil {
		return "", errors.WithMessagef(err, "render template with data error. data:%+v", statistics)
	}
	return buf.String(), nil
}

type XiaohongshuDeliveryStatisticsRender struct {
	Total                        int `json:"total"`
	CollectTimeoutCount          int `json:"collect_timeout_count"`
	CollectWarnCount             int `json:"collect_warn_count"`
	CollectTransportWarnCount    int `json:"collect_transport_warn_count"`
	CollectTransportTimeoutCount int `json:"collect_transport_timeout_count"`
}
