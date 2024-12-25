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
