package alarm

import (
	"context"
	"net/http"
	"testing"

	"PulseCheck/internal/config"
)

func init() {
	config.InitLogger()
}

func TestXiaohongshuStatisticsEmailConsumer_sendEmail(t *testing.T) {
	type args struct {
		from    string
		to      []string
		subject string
		body    string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "sendEmail",
			args:    args{from: "", to: []string{"Azureym@126.com"}, subject: "测试小红书异常提醒", body: "test email body"},
			wantErr: false, // TODO: Check for expected error.
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &XHSStatisticsEmailConsumer{}
			if err := this.sendEmail(tt.args.from, tt.args.to, tt.args.subject, tt.args.body); (err != nil) != tt.wantErr {
				t.Errorf("sendEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestXiaohongshuStatisticsProvider_Provide(t *testing.T) {
	type fields struct {
		httpClient *http.Client
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    <-chan *XHSDeliveryStatistics
		wantErr bool
	}{
		{
			name: "Provide",
			args: args{ctx: context.Background()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := NewXiaohongshuStatisticsProvider()
			got, err := this.Provide(tt.args.ctx)
			if nil != err {
				t.Fatalf("access xiaohongshu logisticas api failed. error:%+v", err)
			}
			t.Logf("get data:%#v", <-got)
		})
	}
}

func TestXiaohongshuStatisticsEmailConsumer_Consume(t *testing.T) {
	type args struct {
		ctx      context.Context
		dataChan <-chan *XHSDeliveryStatistics
	}
	dataChan := make(chan *XHSDeliveryStatistics, 1)
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Execute",
			args:    args{ctx: context.Background(), dataChan: dataChan},
			wantErr: false, // TODO: Check for expected error.
		},
	}
	go func() {
		defer close(dataChan)
		dataChan <- &XHSDeliveryStatistics{
			Total:                        1,
			CollectWarnCount:             100,
			ReturnRejectCount:            99,
			LogisticsStandstillCount:     98,
			CollectTransportWarnCount:    100000,
			CollectTransportTimeoutCount: 22,
			DeliverySignTimeoutCount:     33,
			LogisticsRouteTimeoutCount:   44,
		}
	}()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x := NewXHSStatisticsHandler()
			if err := x.Execute(tt.args.ctx, tt.args.dataChan); (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
