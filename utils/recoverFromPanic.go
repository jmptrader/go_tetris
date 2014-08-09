package utils

import (
	"fmt"
	"runtime/debug"
)

// recover from panic
// log stack information
// run new function
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
			// Noted:
			// should run the function in a new goroutine
			// memory leak otherwise
			go f()
		}
	}
}
