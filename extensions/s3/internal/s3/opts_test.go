package s3

import (
	"testing"
	"time"
)

func TestIsWithRetry(t *testing.T) {
	var opts []Option
	if IsWithRetry(opts...) == true {
		t.Error("IsWithRetry should return false when opts is nil")
	}
	opts = append(opts, WithRetry(10, time.Second))
	if IsWithRetry(opts...) == false {
		t.Error("IsWithRetry should return true")
	}
}
