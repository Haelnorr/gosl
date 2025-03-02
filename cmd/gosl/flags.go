package main

import (
	"flag"
	"strconv"
)

func setupFlags() map[string]string {
	// Parse commandline args
	host := flag.String("host", "", "Override host to listen on")
	port := flag.String("port", "", "Override port to listen on")
	test := flag.Bool("test", false, "Run server in test mode")
	tester := flag.Bool("tester", false, "Run tester function instead of main program")
	dbver := flag.Bool("dbver", false, "Get the version of the database required")
	loglevel := flag.String("loglevel", "", "Set log level")
	logoutput := flag.String("logoutput", "", "Set log destination (file, console or both)")
	flag.Parse()

	// Map the args for easy access
	args := map[string]string{
		"host":      *host,
		"port":      *port,
		"test":      strconv.FormatBool(*test),
		"tester":    strconv.FormatBool(*tester),
		"dbver":     strconv.FormatBool(*dbver),
		"loglevel":  *loglevel,
		"logoutput": *logoutput,
	}
	return args
}
