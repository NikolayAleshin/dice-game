package interfaces

type Context interface {
	Str(key, val string) Context
	Int(key string, val int) Context
	Bool(key string, val bool) Context
	Interface(key string, val interface{}) Context
	Err(err error) Context
	Timestamp() Context
	Logger() Logger
}

type Event interface {
	Msg(msg string)
	Msgf(format string, v ...interface{})
	Interface(key string, val interface{}) Event
	Str(key, val string) Event
	Int(key string, val int) Event
	Bool(key string, val bool) Event
	Err(err error) Event
}

type Logger interface {
	Debug() Event
	Info() Event
	Warn() Event
	Error() Event
	Fatal() Event
	With() Context
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}
