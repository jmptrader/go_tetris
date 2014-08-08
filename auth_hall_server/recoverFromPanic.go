package main

func recoverFromPanic(format string, f func()) {
	if err := recover(); err != nil {
		log.Critical(format+"%v", err)
	}
	go f()
}
