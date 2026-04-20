package response

// Envelope 定义统一响应结构。
type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// OK 返回标准成功响应。
func OK(message string, data any) Envelope {
	return Envelope{
		Code:    0,
		Message: message,
		Data:    data,
	}
}

// Paged 返回标准分页成功响应。
func Paged(message string, page, pageSize int, total int64, items any) Envelope {
	return OK(message, map[string]any{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"items":     items,
	})
}

// Fail 返回标准失败响应。
func Fail(code int, message string) Envelope {
	return Envelope{
		Code:    code,
		Message: message,
	}
}
