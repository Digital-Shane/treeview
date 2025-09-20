package s3

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	cfg "github.com/aws/aws-sdk-go-v2/config"
)

type options struct {
	triggers        []string // triggers message for retry
	profile         string
	retryLapse      time.Duration
	retryMaxAttempt int
	retrying        bool
}

// Option allows parameterizing a function
type Option func(opts *options)

func collectOptions(ops ...Option) *options {
	opts := &options{}
	for _, option := range ops {
		option(opts)
	}
	return opts
}

// collectCfg returns a slice of LoadOptionsFunc from the options.
func collectCfg(ops ...Option) []cfg.LoadOptionsFunc {
	o := collectOptions(ops...)
	var c []cfg.LoadOptionsFunc
	if o.profile != "" {
		c = append(c, cfg.WithSharedConfigProfile(o.profile))
	}
	if o.retrying {
		c = append(c, cfg.WithRetryMaxAttempts(o.retryMaxAttempt), cfg.WithRetryMode(aws.RetryModeStandard))
	}

	return c
}

// IsWithRetry returns true if the option is set to use a retry strategy.
func IsWithRetry(opts ...Option) bool {
	return collectOptions(opts...).retrying
}

// WithProfile indicates the operation should use the profile `p`
func WithProfile(p string) Option {
	return func(op *options) {
		op.profile = p
	}
}

// WithRetry indicates that the operation must use a retry strategy
// with a maximum number of attempts `max` with a minimal delay `lapse`
func WithRetry(maxAtt int, lapse time.Duration, trigger ...string) Option {
	return func(op *options) {
		op.retrying = true
		op.retryMaxAttempt = maxAtt
		op.retryLapse = lapse
		op.triggers = append(op.triggers, trigger...)
	}
}
