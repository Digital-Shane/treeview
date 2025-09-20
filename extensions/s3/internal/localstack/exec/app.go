// Package exec provides a set of system-related functions such a Run that runs an application.
package exec

import (
	"context"
	"os/exec"

	"github.com/pterm/pterm"
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
		pterm.Warning.Printfln("command %s issued error %v %s, retrying", app, err, string(data))
	}
	if !appOpts.verbose {
		return data, err
	}
	s := app1
	for _, a := range appOpts.args {
		s += " " + a
	}
	if appOpts.verbose {
		if err != nil {
			pterm.Error.Printfln("\ncommand %s issued error %v %s", s, err, string(data))
		}
	} else {
		if appOpts.testValue && (err == nil) {
			pterm.Info.Printfln("\ncommand %s did not issue an error", s)
		}
		if !appOpts.testValue && (err != nil) {
			pterm.Warning.Printfln("\ncommand %s issued an error", s)
		}
	}
	return data, err
}
