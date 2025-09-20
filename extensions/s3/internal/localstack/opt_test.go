package localstack

import (
	"testing"
)

func TestWithEnvironmentVariable(t *testing.T) {
	o := collectOptions(WithEnvironmentVariable("Test", "true"))
	if o == nil {
		t.Fatal("Options should not be nil")
	}
	s := o.getEnv()
	if s != `{"Variables":{"Test":"true","USE_LOCALSTACK":"true"}}` {
		t.Error("Options should contain the environment variable set")
	}
}
