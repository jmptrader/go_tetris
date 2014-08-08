package utils

import (
	"fmt"
	"runtime/debug"
)

func RecoverFromPanic(format string, log func(string, ...interface{}), f func()) {
	format += "%v\n%s"
	if err := recover(); err != nil {
		if log != nil {
			log(format, err, debug.Stack())
		} else {
			fmt.Printf(format, err, debug.Stack())
		}
		// run the function in new goroutine
		if f != nil {
			// TODO: not sure if it is better to run the function in a new goroutine or not
			f()
		}
	}
}
