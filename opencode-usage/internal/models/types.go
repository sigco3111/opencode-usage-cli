package models

type Summary struct {
	TotalMessages     int64   `json:"total_messages"`
	UserRequests      int64   `json:"user_requests"`
	InputTokens       int64   `json:"input_tokens"`
	OutputTokens      int64   `json:"output_tokens"`
	ReasoningTokens   int64   `json:"reasoning_tokens"`
	CacheReadTokens   int64   `json:"cache_read_tokens"`
	CacheWriteTokens  int64   `json:"cache_write_tokens"`
	TotalCost         float64 `json:"total_cost"`
	AbortedCount      int64   `json:"aborted_count"`
}

type ModelUsage struct {
	Model       string  `json:"model"`
	Provider    string  `json:"provider"`
	Messages    int64   `json:"messages"`
	InputTokens int64   `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	Cost        float64 `json:"cost"`
}

type DailyUsage struct {
	Date        string  `json:"date"`
	Messages    int64   `json:"messages"`
	InputTokens int64   `json:"input_tokens"`
	OutputTokens int64  `json:"output_tokens"`
	Cost        float64 `json:"cost"`
	CumMessages    int64 `json:"cum_messages,omitempty"`
	CumInputTokens int64 `json:"cum_input_tokens,omitempty"`
	CumOutputTokens int64 `json:"cum_output_tokens,omitempty"`
}

type ProjectUsage struct {
	Project   string `json:"project"`
	Sessions  int64  `json:"sessions"`
	FirstUsed string `json:"first_used"`
	LastUsed  string `json:"last_used"`
}

type HourlyUsage struct {
	Hour        string `json:"hour"`
	Messages    int64  `json:"messages"`
	InputTokens int64  `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

type AgentUsage struct {
	Agent       string `json:"agent"`
	Messages    int64  `json:"messages"`
	InputTokens int64  `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
}

type UsageData struct {
	Summary  *Summary       `json:"summary"`
	Models   []ModelUsage   `json:"models,omitempty"`
	Daily    []DailyUsage   `json:"daily,omitempty"`
	Projects []ProjectUsage `json:"projects,omitempty"`
	Hourly   []HourlyUsage  `json:"hourly,omitempty"`
	Agents   []AgentUsage   `json:"agents,omitempty"`
	DateRange string        `json:"date_range"`
}
