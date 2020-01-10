package importer

import "github.com/mullakhmetov/duplicates-checker/cmd"

// Command ...
type Command struct {
	cmd.CommonOpts
}

// Execute command starts importing data to store
func (c *Command) Execute(args []string) error {
	return nil
}
