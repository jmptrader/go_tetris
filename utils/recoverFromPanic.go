package utils

import "runtime/debug"

func RecoverFromPanic(format string, log func(string, ...interface{}), f func()) {
	if err := recover(); err != nil {
		if log != nil {
			log(format+"%v\n%s", err, debug.Stack())
		}
	}
	if f != nil {
		go f()
	}
}
