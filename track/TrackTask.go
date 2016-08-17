package track

import (
	"github.com/hailongz/kk-go-task/task"
)

type TrackTaskResult struct {
	Result
	Track *Track `json:"track,omitempty"`
}

/**
 * 更新追踪
 */
type TrackTask struct {
	task.Task
	Code   int64 `json:"code"`
	Result TrackTaskResult
}

func (T *TrackTask) API() string {
	return "track/get"
}

func (T *TrackTask) GetResult() interface{} {
	return &T.Result
}
