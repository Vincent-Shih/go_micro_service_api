package response

type Response struct {
	Code    int         `json:"code"`
	TraceID int         `json:"trace_id"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
