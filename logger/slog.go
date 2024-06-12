package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"time"

	"log/slog"
)

// ============ Internal(private) Methods - can only be called from inside this package ==============

type ContextReqId struct{}
type ContextAppId struct{}

func initializeLoggerV3() *slog.Logger {
	enc := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	})
	h := ContextHandler{enc, []any{
		ContextReqId{},
		ContextAppId{},
	}}
	return slog.New(h)
}

// =========== Exposed (public) Methods - can be called from external packages ============

// GetLoggerV3 returns the slog logger object.
func GetLoggerV3() *slog.Logger {
	return initializeLoggerV3()
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

var (
	Log *slog.Logger
)

type SlogLoggerImpl struct {
	Logger *slog.Logger
}

func NewSlogLogger(logger *slog.Logger) Logger {
	return &SlogLoggerImpl{Logger: logger}
}

func InitSlogLogger() *slog.Logger {
	return GetLoggerV3() // Initialize your slog.Logger here
}

func (l *SlogLoggerImpl) log(level slog.Level, msg string, args ...any) {
	var pcs [1]uintptr
	runtime.Callers(3, pcs[:]) // skip [Callers, log, Infof/Errorf]
	r := slog.NewRecord(time.Now(), level, fmt.Sprintf(msg, args...), pcs[0])
	_ = l.Logger.Handler().Handle(context.Background(), r)
}

func (l *SlogLoggerImpl) Infof(msg string, args ...any) {
	l.log(slog.LevelInfo, msg, args...)
}

func (l *SlogLoggerImpl) Errorf(msg string, args ...any) {
	l.log(slog.LevelError, msg, args...)
}

func (l *SlogLoggerImpl) Print(v ...interface{}) {
	l.log(slog.LevelInfo, "Log message: "+fmt.Sprint(v...))
}
