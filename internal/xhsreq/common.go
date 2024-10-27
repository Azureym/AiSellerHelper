package xhsreq

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/spf13/cast"

	"PulseCheck/internal/tools"
)

const (
	cookieFormat = `a1=18f137ccfed9g2hoelwpjqyvnunsiagjjyjd5stox00000135039; webId=a23bf7a341ef68df96c1281831cbd021; gid=yYiyqWSffiVSyYiyqWSSivM9dfjkJxld1EUVVlCAvhI7If88M30kV2888yq28qj80JS4i8i4; customerClientId=688073957365739; x-user-id-zhaoshang.xiaohongshu.com=; ark_worker_plugin_uuid=886235df2bf9422ebe1ea9c73915763c; x-user-id-school.xiaohongshu.com=65d6ecd9e300000000000001; abRequestId=a23bf7a341ef68df96c1281831cbd021; access-token=; sso-type=customer; subsystem=ark; web_session=040069431fa14842cc90b0bc41344b8ee07f5d; webBuild=4.25.1; unread={%22ub%22:%22668bd8750000000025015bc8%22%2C%22ue%22:%22668f74db0000000025002eed%22%2C%22uc%22:33}; acw_tc=2521b57b58b2b5ba5149cbd95cbeb0ba54231d009c1bef6bb34afbd638e02b91; xsecappid=sellercustomer; websectiga=f3d8eaee8a8c63016320d94a1bd00562d516a5417bc43a032a80cbf70f07d5c0; sec_poison_id=ae4e27b9-601e-40f1-b112-4351b82c64f1; customer-sso-sid=68c5173991171136170488473a6fe5c96bf49c16; x-user-id-ark.xiaohongshu.com=5679401044760815fb659cc6; access-token-ark.xiaohongshu.com=customer.ark.{{ . }}; access-token-ark.beta.xiaohongshu.com=customer.ark.{{ . }}; beaker.session.id=1d657de3051687a200364d516edbdc967440edc8gAJ9cQAoWAsAAABhcmstbGlhcy1pZHEBWBgAAAA2NTVmMmY1YjNlYWQ0ZjAwMDE1YWU2YzFxAlgOAAAAcmEtdXNlci1pZC1hcmtxA1gYAAAANjVkNmVjZDllMzAwMDAwMDAwMDAwMDAxcQRYDgAAAF9jcmVhdGlvbl90aW1lcQVHQdmrvHS+JN1YEQAAAHJhLWF1dGgtdG9rZW4tYXJrcQZYQQAAADM2MTZjOWM1YzhlNDQyYjBhZjI1NWIzOGIyZjE4YWM5LTBjODNiYzEzNjg3YzQ4ZTM5OGFlYTkyYjY0ZjBhYzA4cQdYAwAAAF9pZHEIWCAAAAA4MzhiZWNjZjE4ZWY0YTQ5OTZjZTQ1MWUxZTJlZmQwN3EJWA4AAABfYWNjZXNzZWRfdGltZXEKR0HZq7x0viTddS4=`
)

func generateCookie(token string) string {
	t, b := template.New("cookieTemplate"), new(strings.Builder)
	err := template.Must(t.Parse(cookieFormat)).Execute(b, token)
	if err != nil {
		log.Fatalf("could not parse cookie. %s. error:%+v", cookieFormat, err)
	}
	return b.String()
}

func ResetCommonHeaders(ctx context.Context, header http.Header) error {
	token := tools.WithdrawXHSToken(ctx)
	cookie := generateCookie(token)
	header.Set("cookie", cookie)

	header.Set("accept", "application/json, text/plain, */*")
	header.Set("accept-language", "en")
	header.Set("content-type", "application/json")
	header.Set("cache-control", "no-cache")
	header.Set("dnt", "1")
	header.Set("origin", "https://ark.xiaohongshu.com")
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
	slog.Debug(spew.Sprintf("header reset. header:%#v", header))
	return nil
}

func AddRefererHeader(ctx context.Context, header http.Header, referer string) {
	header.Set("referer", referer)
}

const (
	separator = "."
)

type Path string

func (this Path) Index(i int) Path {
	join := strings.Join([]string{string(this), strconv.Itoa(i)}, separator)
	return Path(join)
}
func (this Path) Join(s Path) Path {
	join := strings.Join([]string{string(this), string(s)}, separator)
	return Path(join)
}

func (this Path) String() string {
	return string(this)
}

func FromString(s string) Path {
	return Path(s)
}

func JsonDataPath(nodes ...Path) string {
	stringNode := make([]string, len(nodes))
	for _, path := range nodes {
		stringNode = append(stringNode, string(path))
	}
	if len(stringNode) == 1 {
		return stringNode[0]
	}

	return strings.Join(stringNode, ".")
}

const (
	XiaohongshuDomain = "ark.xiaohongshu.com"
	DifyDomain        = "api.dify.ai"
)
