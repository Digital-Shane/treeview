package config

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/google/go-cmp/cmp"
)

func TestMain(m *testing.M) {
	_ = os.Setenv(Env, "true")
	defer func() { _ = os.Unsetenv(Env) }()
	m.Run()
}

func TestNewConfig(t *testing.T) {
	c, err := NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cmp.Equal(aws.Config{}, c) {
		t.Error("NewConfig should not return default config")
	}
	_, err = NewConfig()
	if err != nil {
		t.Fatal(err)
	}
	// Test cache
	_, ok := _cached.get()
	if ok == false {
		t.Fatal("NewConfig() should return a cached config")
	}
	ClearCached()
	_, ok = _cached.get()
	if ok == true {
		t.Fatal("NewConfig() should not return a cached config")
	}
}
