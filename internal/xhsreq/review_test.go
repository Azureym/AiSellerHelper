package xhsreq

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"PulseCheck/internal/tools"
)

func TestReviewManager_GetReviews(t *testing.T) {
	type fields struct {
		httpClient *http.Client
	}
	type args struct {
		ctx   context.Context
		param *ReviewSearchParam
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantReviews []*Review
		wantErr     bool
	}{
		{
			name: "test",
			fields: fields{
				httpClient: tools.NewHttpsClient(nil, XiaohongshuDomain),
			},
			args: args{
				ctx: tools.AppendXHSToken(context.Background(), "AT-68c517428216101088487948dfhuvp629ni26cvc"),
				param: &ReviewSearchParam{
					OrderID: "P744461199153402613",
				},
			},
		},

		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			this := &ReviewManager{
				httpClient: tt.fields.httpClient,
			}
			gotReviews, err := this.GetReviews(tt.args.ctx, tt.args.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReviews() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReviews, tt.wantReviews) {
				t.Errorf("GetReviews() gotReviews = %v, want %v", gotReviews, tt.wantReviews)
			}
		})
	}
}
