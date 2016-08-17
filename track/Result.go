package track

type IResultTask interface {
	GetResult() interface{}
}

type Result struct {
	Errno  int    `json:"errno,omitempty"`
	Errmsg string `json:"errmsg,omitempty"`
}

const ERRNO_TRACK = 0x2000

/**
 * 未找到跟踪码
 */
const ERRNO_NOT_FOUND_CODE = ERRNO_TRACK + 1

/**
 * 未找到IP
 */
const ERRNO_NOT_FOUND_IP = ERRNO_TRACK + 2
