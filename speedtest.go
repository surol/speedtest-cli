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

	switch {
	case opts.Help:
		usage()
		return
	case opts.Version:
		version()
		return
	}

	client := speedtest.NewClient(opts)

	serversChan := make(chan speedtest.ServersRef)
	client.AllServers(serversChan)

	serversRef := <- serversChan
	if serversRef.Error != nil {
		log.Fatal(serversRef.Error);
	}
	servers := serversRef.Servers

	if opts.List {
		fmt.Println(servers)
		return
	}

	if len(opts.Server) == 0 {
		servers.Truncate(5)
	}
}
