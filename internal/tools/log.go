package tools

import (
	"context"
	"fmt"
	"net/http"
)

func LogFromContext(ctx context.Context, format string, args ...interface{}) {
	writer := WithdrawWriter(ctx)
	if nil == writer {
		return
	}
	defer func() {
		flusher, ok := writer.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	}()
	fmt.Fprintf(writer, format+"\n", args...)
}
