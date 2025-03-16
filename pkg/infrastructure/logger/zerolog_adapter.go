package logger

import (
	"dice-game/pkg/domain/interfaces"
	"github.com/rs/zerolog"
	"os"
)

type ZerologEvent struct {
	event *zerolog.Event
}

func (e *ZerologEvent) Msg(msg string) {
	e.event.Msg(msg)
}

func (e *ZerologEvent) Msgf(format string, v ...interface{}) {
	e.event.Msgf(format, v...)
}

func (e *ZerologEvent) Interface(key string, val interface{}) interfaces.Event {
	e.event.Interface(key, val)
	return e
}

func (e *ZerologEvent) Str(key, val string) interfaces.Event {
	e.event.Str(key, val)
	return e
}

func (e *ZerologEvent) Int(key string, val int) interfaces.Event {
	e.event.Int(key, val)
	return e
}

func (e *ZerologEvent) Bool(key string, val bool) interfaces.Event {
	e.event.Bool(key, val)
	return e
}

func (e *ZerologEvent) Err(err error) interfaces.Event {
	e.event.Err(err)
	return e
}

type ZerologContext struct {
	ctx    zerolog.Context
	parent *ZerologAdapter
}

func (c *ZerologContext) Str(key, val string) interfaces.Context {
	c.ctx = c.ctx.Str(key, val)
	return c
}

func (c *ZerologContext) Int(key string, val int) interfaces.Context {
	c.ctx = c.ctx.Int(key, val)
	return c
}

func (c *ZerologContext) Bool(key string, val bool) interfaces.Context {
	c.ctx = c.ctx.Bool(key, val)
	return c
}

func (c *ZerologContext) Interface(key string, val interface{}) interfaces.Context {
	c.ctx = c.ctx.Interface(key, val)
	return c
}

func (c *ZerologContext) Err(err error) interfaces.Context {
	c.ctx = c.ctx.Err(err)
	return c
}

func (c *ZerologContext) Timestamp() interfaces.Context {
	c.ctx = c.ctx.Timestamp()
	return c
}

func (c *ZerologContext) Logger() interfaces.Logger {
	return &ZerologAdapter{logger: c.ctx.Logger()}
}

type ZerologAdapter struct {
	logger zerolog.Logger
}

func NewZerologAdapter(level string, isJSON bool) *ZerologAdapter {
	var zerologLevel zerolog.Level

	switch level {
	case "debug":
		zerologLevel = zerolog.DebugLevel
	case "info":
		zerologLevel = zerolog.InfoLevel
	case "warn":
		zerologLevel = zerolog.WarnLevel
	case "error":
		zerologLevel = zerolog.ErrorLevel
	default:
		zerologLevel = zerolog.InfoLevel
	}

	var logger zerolog.Logger

	if isJSON {
		logger = zerolog.New(os.Stdout).
			Level(zerologLevel).
			With().
			Timestamp().
			Logger()
	} else {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02T15:04:05Z07:00",
		}

		logger = zerolog.New(output).
			Level(zerologLevel).
			With().
			Timestamp().
			Logger()
	}

	return &ZerologAdapter{logger: logger}
}

func (z *ZerologAdapter) With() interfaces.Context {
	return &ZerologContext{
		ctx:    z.logger.With(),
		parent: z,
	}
}

func (z *ZerologAdapter) Debug() interfaces.Event {
	return &ZerologEvent{event: z.logger.Debug()}
}

func (z *ZerologAdapter) Info() interfaces.Event {
	return &ZerologEvent{event: z.logger.Info()}
}

func (z *ZerologAdapter) Warn() interfaces.Event {
	return &ZerologEvent{event: z.logger.Warn()}
}

func (z *ZerologAdapter) Error() interfaces.Event {
	return &ZerologEvent{event: z.logger.Error()}
}

func (z *ZerologAdapter) Fatal() interfaces.Event {
	return &ZerologEvent{event: z.logger.Fatal()}
}

func (z *ZerologAdapter) WithField(key string, value interface{}) interfaces.Logger {
	newLogger := z.logger.With().Interface(key, value).Logger()
	return &ZerologAdapter{logger: newLogger}
}

func (z *ZerologAdapter) WithFields(fields map[string]interface{}) interfaces.Logger {
	ctx := z.logger.With()
	for k, v := range fields {
		ctx = ctx.Interface(k, v)
	}
	newLogger := ctx.Logger()
	return &ZerologAdapter{logger: newLogger}
}

func (z *ZerologAdapter) WithError(err error) interfaces.Logger {
	newLogger := z.logger.With().Err(err).Logger()
	return &ZerologAdapter{logger: newLogger}
}
