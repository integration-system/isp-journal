package log

type Option func(l *defaultLogger)

func WithAfterRotation(callback func(prevFile LogFile)) Option {
	return func(l *defaultLogger) {
		l.afterRotation = callback
	}
}
