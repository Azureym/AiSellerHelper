package xhsreq

import (
	"context"
	"net/http"
	"testing"

	"PulseCheck/internal/tools"
)

func TestReviewReply_Reply(t *testing.T) {
	type fields struct {
		httpClient *http.Client
	}
	type args struct {
		ctx   context.Context
		param *ReviewReplyParam
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test",
			fields: fields{
				httpClient: tools.NewHttpsClient(nil, XiaohongshuDomain),
			},
			args: args{
				ctx: tools.AppendXHSToken(context.Background(), "AT-68c517428216101088487948dfhuvp629ni26cvc"),
				param: &ReviewReplyParam{
					ReviewIds:    []string{"414486735538025535"},
					ReplyContent: "宝子太有眼光啦～谢谢支持！😚",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &ReviewReply{
				httpClient: tt.fields.httpClient,
			}
			if err := this.Reply(tt.args.ctx, tt.args.param); (err != nil) != tt.wantErr {
				t.Errorf("Reply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
