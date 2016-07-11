package main

import (
	"bitbucket.org/lorus/speedtest-cli/speedtest"
	"fmt"
	"os"
	"flag"
	"log"
)

func version() {
	fmt.Print(speedtest.Version)
}

func usage() {
	fmt.Fprint(os.Stderr, "Command line interface for testing internet bandwidth using speedtest.net.\n\n")
	flag.PrintDefaults()
}

func main() {
	opts := speedtest.ParseOpts()

	if opts.Version {
		version()
		return
	}

	if opts.Help {
		usage();
		return
	}

	client := speedtest.NewClient(opts)

	_, err := client.Config()
	if err != nil {
		log.Fatal(err)
	}
}
