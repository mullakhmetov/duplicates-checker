package cmd

// Interface for commands
type Interface interface {
	SetCommon(commonOpts CommonOpts)
	Execute(args []string) error
}

// CommonOpts keeps common options, shared across all commands
type CommonOpts struct {
	Revision   string
	BoltDBName string
	Dbg        bool
}

// SetCommon sets common option fields
// The method called by main for each command
func (c *CommonOpts) SetCommon(commonOpts CommonOpts) {
	c.Revision = commonOpts.Revision
	c.BoltDBName = commonOpts.BoltDBName
	c.Dbg = commonOpts.Dbg
}
