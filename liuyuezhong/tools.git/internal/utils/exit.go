package utils

import "sync/atomic"

var exit int32

func InitExitSign() {
	ctx := OnExit()
	go func() {
		<-ctx.Done()
		atomic.StoreInt32(&exit, 1)
	}()
}

func IsExit() bool {
	return atomic.LoadInt32(&exit) == 1
}
