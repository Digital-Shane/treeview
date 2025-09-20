// Package config prepares the AWS configuration that will be used by the AWS services.  It handles the localstack
// configuration. It provides some means to generate errors useful in unitary tests
package config

import (
	"context"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/pkg/errors"
)

const (
	// Env is the environment variable name for the localstack use.  If set, the aws library uses localstack.
	// Its value is irrelevant
	Env = "USE_LOCALSTACK"
	// EnvFaulty is the environment variable name for the faulty configuration use.  If set, the aws library uses
	// a faulty configuration.
	EnvFaulty = "USE_FAULTY_LOCALSTACK"
	// AwsProfile is the environment variable name for the aws profile to use.  If set, the aws library uses the
	// profile.
	AwsProfile = "GO_AWS_PROFILE" // Trying to avoid collision. Terraform uses AWS_PROFILE.

	_cFaultyRegion = "faulty-region-1"
)

// _cached is the cached configuration to avoid throttling when calling too often the configuration.
var _cached cache

// NewConfig returns a configuration with the given options.  If there are no options, it is the default configuration.
// The supported options are `WithLocalStack`.
func NewConfig(opts ...config.LoadOptionsFunc) (cfg aws.Config, err error) {
	var opa []func(loadOptions *config.LoadOptions) error
	const cDefaultRegion = "us-west-2" // The default region used if the configuration does not define a region.
	for _, opt := range opts {
		opa = append(opa, opt)
	}
	if os.Getenv(AwsProfile) != "" {
		opa = append(opa, config.WithSharedConfigProfile(os.Getenv(AwsProfile)))
	}
	cfg, err = config.LoadDefaultConfig(context.Background(), opa...)
	if err != nil {
		return aws.Config{}, err
	}
	switch cfg.Region {
	case "": // attributes automatically the default region
		cfg.Region = cDefaultRegion
	case _cFaultyRegion: // requests an error for test purpose.
		err = errors.New("faulty configuration")
		return aws.Config{}, err
	}
	if os.Getenv(EnvFaulty) != "" {
		return cfg, errors.New("faulty configuration")
	}
	return setCached(cfg, err)
}

// ClearCached clears the cached configuration.
func ClearCached() {
	_cached.clear()
}

// SetProfile sets the AWS profile to use by `go-aws`.  If the profile is empty, it uses the default profile.
// It returns the current profile used by `go-aws` and an error if any.
func SetProfile(profile string) (string, error) {
	if profile == "" {
		profile = "default"
	}
	old, ok := os.LookupEnv(AwsProfile)
	if !ok {
		old = "default"
	}
	return old, os.Setenv(AwsProfile, profile)
}

// ---------------------------------

func setCached(cfg aws.Config, err error) (aws.Config, error) {
	if err != nil {
		return aws.Config{}, err
	}
	err = _cached.Set(cfg)
	if err != nil {
		return aws.Config{}, err
	}
	return cfg, nil
}

type cache struct {
	cfg   aws.Config
	crd   aws.Credentials
	lock  sync.RWMutex
	valid bool
}

func (c *cache) clear() {
	c.lock.Lock()
	c.valid = false
	c.lock.Unlock()
}

func (c *cache) get() (aws.Config, bool) {
	if !c.valid {
		return aws.Config{}, false
	}
	c.lock.RLock()
	defer c.lock.RUnlock()
	if !c.crd.CanExpire {
		return c.cfg, true
	}
	if !c.crd.Expired() {
		return c.cfg, true
	}
	c.valid = false
	return aws.Config{}, false
}

func (c *cache) Set(cfg aws.Config) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.cfg = cfg
	crd, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil {
		c.valid = false
		return err
	}
	c.crd = crd
	c.valid = true
	return nil
}
