// Package exec provides a set of system-related functions such a Run that runs an application.
package exec

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes the application `app` using the parameters defines by `options`.  It returns the output of the
// application, and an error if any.
func Run(app string, opts ...Option) ([]byte, error) {
	appOpts := collectOptions(opts...)
	var data []byte
	var err error
	availableRetries := appOpts.retryNumber
	if !appOpts.retry {
		availableRetries = 0
	}
	app1 := app
	for {
		cmd := exec.CommandContext(context.Background(), app1, appOpts.args...)
		if !appOpts.verbose {
			data, err = cmd.Output()
		} else {
			data, err = cmd.CombinedOutput()
		}
		if err == nil {
			break
		}
		if availableRetries == 0 {
			break
		}
		availableRetries--
		fmt.Printf("Warn: command %s issued error %v %s, retrying\n", app, err, string(data))
	}
	if !appOpts.verbose {
		return data, err
	}
	var s strings.Builder
	s.WriteString(app1)
	for _, a := range appOpts.args {
		s.WriteString(" " + a)
	}
	if appOpts.verbose {
		if err != nil {
			fmt.Printf("\nError: command %s issued error %v %s\n", s.String(), err, string(data))
		}
	} else {
		if appOpts.testValue && (err == nil) {
			fmt.Printf("\ncommand %s did not issue an error\n", s.String())
		}
		if !appOpts.testValue && (err != nil) {
			fmt.Printf("\nWarn: command %s issued an error\n", s.String())
		}
	}
	return data, err
}
