package track

import (
	"github.com/hailongz/kk-go-task/task"
)

type TrackSetTaskResult struct {
	Result
	Code int64 `json:"code,omitempty"`
}

/**
 * 更新追踪
 */
type TrackSetTask struct {
	task.Task
	Code      int64   `json:"code"`
	IP        string  `json:"ip"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Result    TrackSetTaskResult
}

func (T *TrackSetTask) API() string {
	return "track/set"
}

func (T *TrackSetTask) GetResult() interface{} {
	return &T.Result
}
