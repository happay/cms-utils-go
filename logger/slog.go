package logger

import (
	"context"
	"os"
	"sync"

	"golang.org/x/exp/slog"
)

// ============ Internal(private) Methods - can only be called from inside this package ==============

type ContextReqId struct{}
type ContextAppId struct{}

var sLog *slog.Logger

func initializeLoggerV3() {
	enc := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})
	h := ContextHandler{enc, []any{
		ContextReqId{},
		ContextAppId{},
	}}
	sLog = slog.New(h)
}

var sLogInit sync.Once

// =========== Exposed (public) Methods - can be called from external packages ============

// GetLogger returns the logrus logger object. It takes three input parameters.
// - logPrefix - it is a string used as Prefix on each log line
// - logPath - absolute path of the log file where the logs will be written
// - appName - It is app Name, from which service this function is being called to route the log to a specific Graylog stream.
func GetLoggerV3() *slog.Logger {
	sLogInit.Do(func() {
		initializeLoggerV3()
	})
	return sLog
}

type ContextHandler struct {
	slog.Handler
	keys []any
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(h.observe(ctx)...)
	return h.Handler.Handle(ctx, r)
}

func (h ContextHandler) observe(ctx context.Context) (as []slog.Attr) {
	for _, k := range h.keys {
		a, ok := ctx.Value(k).(slog.Attr)
		if !ok {
			continue
		}
		a.Value = a.Value.Resolve()
		as = append(as, a)
	}
	return
}
