package logger

import "sync"

var (
	gl    Logger = &NopLogger{}
	mutex sync.RWMutex
)

func SetGlobalLogger(l Logger) {
	mutex.Lock()
	defer mutex.Unlock()
	gl = l
}

func L() Logger {
	mutex.RLock()
	ret := gl
	mutex.RUnlock()
	return ret
}
