package internal

import "time"

type ManualRequest struct {
	Governor string `json:"governor"`
	Profile  string `json:"profile"`
}

type TimerRequest struct {
	Duration time.Duration `json:"duration"`
	Governor string        `json:"governor"`
	Profile  string        `json:"profile"`
}

type AppErr struct {
	Msg        string `json:"msg"`
	StatusCode int    `json:"status_code"`
	Err        error  `json:"-"`
}

func (a AppErr) Error() string {
	return a.Err.Error()
}
