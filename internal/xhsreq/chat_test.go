package xhsreq

import (
	"context"
	"net/http"
	"testing"

	"PulseCheck/internal/tools"
)

func TestXHSReviewChat_Interact(t *testing.T) {
	type fields struct {
		httpClient *http.Client
	}
	type args struct {
		ctx   context.Context
		param *XHSReviewChatParam
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				httpClient: tools.NewHttpsClient(nil, DifyDomain),
			},
			args: args{
				ctx: tools.AppendXHSToken(context.Background(), "AT-68c517428216101088487948dfhuvp629ni26cvc"),
				param: &XHSReviewChatParam{
					ItemId:        "6564c049474aad0001c7641a",
					ItemInfo:      "通过裤夹可以将裤子挂起来收纳到衣柜中,简单方便,整齐,省空间",
					ReviewContent: "不要买！ 质量很差 ！跟我之前买的一点也不一样！合住都对不上！",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &XHSReviewChat{
				httpClient: tt.fields.httpClient,
			}
			got, err := this.Interact(tt.args.ctx, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("Interact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Interact() got = %v, want %v", got, tt.want)
			}
		})
	}
}
