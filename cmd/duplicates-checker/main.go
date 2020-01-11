package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/mullakhmetov/duplicates-checker/cmd"
	"github.com/mullakhmetov/duplicates-checker/cmd/importer"
	"github.com/mullakhmetov/duplicates-checker/cmd/rest"
)

// sets at buildtime via ldflags
var revision = "unknown"

type opts struct {
	Rest     rest.Command     `command:"server"`
	Importer importer.Command `command:"import"`
}

func main() {
	fmt.Printf("duplicates-checker revision: %s\n", revision)

	// parse args and decide what should we do
	var opts opts
	p := flags.NewParser(&opts, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		c := command.(cmd.Interface)
		c.SetCommon(cmd.CommonOpts{
			Revision: revision,
		})
		err := c.Execute(args)
		if err != nil {
			log.Printf("[ERROR] failed with %+v", err)
		}
		return err
	}

	// unknown command
	if _, err := p.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

}
