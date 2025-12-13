package logging

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rs/zerolog"
)

type ZerologHandler struct {
	logger zerolog.Logger
	level  slog.Level
	attrs  []slog.Attr
	groups []string
}

func NewZerologHandler(logger zerolog.Logger, level slog.Level) *ZerologHandler {
	return &ZerologHandler{
		logger: logger,
		level:  level,
	}
}

func (h *ZerologHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *ZerologHandler) Handle(_ context.Context, record slog.Record) error {
	event := h.eventFor(record.Level)
	if event == nil {
		return nil
	}

	for _, attr := range h.attrs {
		h.addAttr(event, h.key(attr.Key), attr.Value)
	}
	record.Attrs(func(attr slog.Attr) bool {
		h.addAttr(event, h.key(attr.Key), attr.Value)
		return true
	})

	event.Msg(record.Message)
	return nil
}

func (h *ZerologHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	cp := h.clone()
	cp.attrs = append(cp.attrs, attrs...)
	return cp
}

func (h *ZerologHandler) WithGroup(name string) slog.Handler {
	if strings.TrimSpace(name) == "" {
		return h
	}
	cp := h.clone()
	cp.groups = append(cp.groups, name)
	return cp
}

func (h *ZerologHandler) clone() *ZerologHandler {
	cp := *h
	cp.attrs = append([]slog.Attr(nil), h.attrs...)
	cp.groups = append([]string(nil), h.groups...)
	return &cp
}

func (h *ZerologHandler) key(k string) string {
	k = strings.TrimSpace(k)
	if k == "" {
		return ""
	}
	if len(h.groups) == 0 {
		return k
	}
	return strings.Join(h.groups, ".") + "." + k
}

func (h *ZerologHandler) eventFor(level slog.Level) *zerolog.Event {
	switch {
	case level >= slog.LevelError:
		return h.logger.Error()
	case level >= slog.LevelWarn:
		return h.logger.Warn()
	case level >= slog.LevelInfo:
		return h.logger.Info()
	default:
		return h.logger.Debug()
	}
}

func (h *ZerologHandler) addAttr(event *zerolog.Event, key string, value slog.Value) {
	if event == nil || strings.TrimSpace(key) == "" {
		return
	}

	value = value.Resolve()

	switch value.Kind() {
	case slog.KindGroup:
		for _, groupAttr := range value.Group() {
			groupKey := h.key(key + "." + groupAttr.Key)
			h.addAttr(event, groupKey, groupAttr.Value.Resolve())
		}
	case slog.KindString:
		event.Str(key, value.String())
	case slog.KindBool:
		event.Bool(key, value.Bool())
	case slog.KindInt64:
		event.Int64(key, value.Int64())
	case slog.KindUint64:
		event.Uint64(key, value.Uint64())
	case slog.KindFloat64:
		event.Float64(key, value.Float64())
	case slog.KindDuration:
		event.Dur(key, value.Duration())
	case slog.KindTime:
		event.Time(key, value.Time())
	default:
		anyValue := value.Any()
		if err, ok := anyValue.(error); ok {
			event.AnErr(key, err)
			return
		}
		if stringer, ok := anyValue.(fmt.Stringer); ok {
			event.Str(key, stringer.String())
			return
		}
		event.Interface(key, anyValue)
	}
}

