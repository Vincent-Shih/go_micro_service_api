// https://github.com/gin-gonic/gin/blob/ce20f107f5dc498ec7489d7739541a25dcd48463/context_test.go#L1747-L1765
package helper

import "net/http/httptest"

type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func (r *TestResponseRecorder) CloseClient() {
	r.closeChannel <- true
}

func (r *TestResponseRecorder) Done() {
	<-r.closeChannel
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}
