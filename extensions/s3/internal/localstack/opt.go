package localstack

import "encoding/json"

type envM map[string]string

type options struct {
	v            envM
	noErrIfExist bool
}

// Option allows parameterizing function
type Option func(opts *options)

func collectOptions(opts ...Option) *options {
	ops := &options{}
	for _, option := range opts {
		option(ops)
	}
	if ops.v == nil {
		ops.v = make(envM)
	}
	ops.v["USE_LOCALSTACK"] = "true"
	return ops
}

// WithEnvironmentVariable allows to set environment variables
func WithEnvironmentVariable(key, value string) Option {
	return func(op *options) {
		if op.v == nil {
			op.v = make(envM)
			op.v["USE_LOCALSTACK"] = "true"
		}
		op.v[key] = value
	}
}

// WithNoErrorIfExist filters out the error if the bucket to be created already exists.
func WithNoErrorIfExist() Option {
	return func(op *options) {
		op.noErrIfExist = true
	}
}
func (o options) getEnv() string {
	type variables struct {
		Variables envM `json:"Variables"`
	}
	v := variables{Variables: o.v}
	b, _ := json.Marshal(v)
	return string(b)
}
