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
	Rest     rest.Command     `command:"server" description:"Starts REST server"`
	Importer importer.Command `command:"import" description:"Starts randomly generated dataset loading. See import command help for details"`

	BoltDBName string `long:"boltdbname" env:"CHECKER_BOLT_DB_NAME" default:"my.db" description:"boltdb db name"`
	Dbg        bool   `long:"dbg" env:"DEBUG" description:"debug mode"`
}

func main() {
	fmt.Printf("duplicates-checker revision: %s\n", revision)

	// parse args and decide what should we do
	var opts opts
	p := flags.NewParser(&opts, flags.Default)
	p.CommandHandler = func(command flags.Commander, args []string) error {
		c := command.(cmd.Interface)
		c.SetCommon(cmd.CommonOpts{
			Revision:   revision,
			BoltDBName: opts.BoltDBName,
			Dbg:        opts.Dbg,
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
