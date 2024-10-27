package config

import (
	"bufio"
	"log/slog"
	"os"
	"path"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

var once sync.Once
var out atomic.Pointer[bufio.Writer]

func InitLogger() {
	once.Do(func() {
		slog.Info("start to initialize log.")
		defer slog.Info("log initialized.")
		opts := &slog.HandlerOptions{
			// Use the ReplaceAttr function on the handler options
			// to be able to replace any single attribute in the log output
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.SourceKey {
					s := a.Value.Any().(*slog.Source)
					s.File = path.Base(s.File)
					return a
				}
				// check that we are handling the time key
				if a.Key == slog.TimeKey {
					t := a.Value.Time()

					// change the value from a time.Time to a String
					// where the string has the correct time format.
					a.Value = slog.StringValue(t.Format(time.DateTime))
					return a
				}
				return a
			},
			AddSource: true,
		}
		writer := bufio.NewWriterSize(os.Stdout, 0)
		out.Store(writer)
		logger := slog.New(slog.NewTextHandler(os.Stdout, opts))
		slog.SetDefault(logger)
	})
}

func CloseLogger() error {
	defer slog.Info("log has been closed.")
	load := out.Load()
	if nil == load {
		return errors.New("load logger failed")
	}
	err := load.Flush()
	if err != nil {
		return errors.WithMessagef(err, "flush logger failed")
	}
	return nil
}
