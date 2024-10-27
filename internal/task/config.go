package task

import "net/url"

type CrawlerExecutorConfig struct {
	//ListenerConfigs []*ListenerConfig
}
type ListenerConfig[T comparable] struct {
	Domain string
	// URL
	// request, request data structure
	Request ListenerRequest
	// response , response data structure
	Response ListenerResponse[T]
	// email, response data converter to VO, email template file path
	Email EmailSender
}

type HttpMethod string

func (this HttpMethod) String() string {
	return string(this)
}

const (
	POST HttpMethod = "POST"
	GET  HttpMethod = "GET"
)

type ListenerRequest struct {
	Url     *url.URL
	Method  HttpMethod
	Headers map[string]string
	Body    []byte
}

type JsonType int

type ListenerResponse[T comparable] struct {
	Fields map[string]struct{}
}

type EmailSender struct {
	Domain string
	Host   string
	Port   string
	From   string
	Passwd string
	To     []string
}
