package timer

import (
	"time"
)

type Timer interface {
	Stop() bool
	Reset(d time.Duration) bool
}
