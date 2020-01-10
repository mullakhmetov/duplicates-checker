package cmd

type Interface interface {
	Execute(args []string) error
	SetCommon()
}

type CommonOpts struct {
	Revision string
}
