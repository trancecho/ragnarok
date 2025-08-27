package fastgpt

type Request struct {
	ChatID    string                 `json:"chatId"`
	Stream    bool                   `json:"stream"`
	Detail    bool                   `json:"detail"` // 一般false
	Variables map[string]interface{} `json:"variables"`
	Messages  []Message              `json:"messages"`
}

type Response struct {
	Id      string `json:"id"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
