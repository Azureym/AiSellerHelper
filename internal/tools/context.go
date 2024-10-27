package tools

import (
	"context"
	"io"
	"log/slog"
)

const (
	XHSToken = "XHS-Token"
	Writer   = "writer"
)

func NewValueContext(data map[string]string) context.Context {
	ctx := context.Background()
	for key, value := range data {
		ctx = context.WithValue(ctx, key, value)
	}
	return ctx
}

func AppendXHSToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, XHSToken, token)
}

func WithdrawXHSToken(ctx context.Context) string {
	token, ok := ctx.Value(XHSToken).(string)
	if !ok {
		slog.Error("can not get XHS token from context")
		panic("can not get XHS token from context")
	}
	return token
}

func AppendWriter(ctx context.Context, writeCloser io.Writer) context.Context {
	return context.WithValue(ctx, Writer, writeCloser)
}

func WithdrawWriter(ctx context.Context) io.Writer {
	writer, ok := ctx.Value(Writer).(io.Writer)
	if !ok {
		return nil
	}
	return writer
}
