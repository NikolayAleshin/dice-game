package logger

import (
	"dice-game/pkg/domain/interfaces"
	"fmt"
)

type MockContext struct {
	fields map[string]interface{}
	logger *MockLogger
}

func (c *MockContext) Str(key, val string) interfaces.Context {
	c.fields[key] = val
	return c
}

func (c *MockContext) Int(key string, val int) interfaces.Context {
	c.fields[key] = val
	return c
}

func (c *MockContext) Bool(key string, val bool) interfaces.Context {
	c.fields[key] = val
	return c
}

func (c *MockContext) Interface(key string, val interface{}) interfaces.Context {
	c.fields[key] = val
	return c
}

func (c *MockContext) Err(err error) interfaces.Context {
	c.fields["error"] = err
	return c
}

func (c *MockContext) Timestamp() interfaces.Context {
	c.fields["timestamp"] = true
	return c
}

func (c *MockContext) Logger() interfaces.Logger {
	return c.logger
}

type MockEvent struct {
	logger *MockLogger
	fields map[string]interface{}
	level  string
}

func (e *MockEvent) Msg(msg string) {
	message := fmt.Sprintf("[%s] %s", e.level, msg)
	switch e.level {
	case "debug":
		e.logger.DebugMessages = append(e.logger.DebugMessages, message)
	case "info":
		e.logger.InfoMessages = append(e.logger.InfoMessages, message)
	case "warn":
		e.logger.WarnMessages = append(e.logger.WarnMessages, message)
	case "error":
		e.logger.ErrorMessages = append(e.logger.ErrorMessages, message)
	case "fatal":
		e.logger.FatalMessages = append(e.logger.FatalMessages, message)
	}
}

func (e *MockEvent) Msgf(format string, v ...interface{}) {
	e.Msg(fmt.Sprintf(format, v...))
}

func (e *MockEvent) Interface(key string, val interface{}) interfaces.Event {
	e.fields[key] = val
	return e
}

func (e *MockEvent) Str(key, val string) interfaces.Event {
	e.fields[key] = val
	return e
}

func (e *MockEvent) Int(key string, val int) interfaces.Event {
	e.fields[key] = val
	return e
}

func (e *MockEvent) Bool(key string, val bool) interfaces.Event {
	e.fields[key] = val
	return e
}

func (e *MockEvent) Err(err error) interfaces.Event {
	e.fields["error"] = err
	return e
}

type MockLogger struct {
	DebugMessages []string
	InfoMessages  []string
	WarnMessages  []string
	ErrorMessages []string
	FatalMessages []string
}

func NewMockLogger() *MockLogger {
	return &MockLogger{
		DebugMessages: make([]string, 0),
		InfoMessages:  make([]string, 0),
		WarnMessages:  make([]string, 0),
		ErrorMessages: make([]string, 0),
		FatalMessages: make([]string, 0),
	}
}

func (m *MockLogger) With() interfaces.Context {
	return &MockContext{
		fields: make(map[string]interface{}),
		logger: m,
	}
}

func (m *MockLogger) Debug() interfaces.Event {
	return &MockEvent{logger: m, fields: make(map[string]interface{}), level: "debug"}
}

func (m *MockLogger) Info() interfaces.Event {
	return &MockEvent{logger: m, fields: make(map[string]interface{}), level: "info"}
}

func (m *MockLogger) Warn() interfaces.Event {
	return &MockEvent{logger: m, fields: make(map[string]interface{}), level: "warn"}
}

func (m *MockLogger) Error() interfaces.Event {
	return &MockEvent{logger: m, fields: make(map[string]interface{}), level: "error"}
}

func (m *MockLogger) Fatal() interfaces.Event {
	return &MockEvent{logger: m, fields: make(map[string]interface{}), level: "fatal"}
}

func (m *MockLogger) WithField(key string, value interface{}) interfaces.Logger {
	return m
}

func (m *MockLogger) WithFields(fields map[string]interface{}) interfaces.Logger {
	return m
}

func (m *MockLogger) WithError(err error) interfaces.Logger {
	return m
}
