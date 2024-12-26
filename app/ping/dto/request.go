package dto

// PingRequest ping请求参数
type PingRequest struct {
	Echo string `json:"echo" form:"echo"` // 可选的回声参数
}
