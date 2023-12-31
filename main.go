package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// config filename
const configFile = "config.json"

// Timeout is the amount of time the server will wait for requests to finish during shutdown
const Timeout = 10 * time.Second

// Configuration is the struct that gets filled by reading config.json JSON file
type Configuration struct {
	VerboseOutput       bool   `json:"verbose"`
	InterfaceAndPort    string `json:"interfaceAndPort"`
	ResponseFile        string `json:"responseFile"`
	ResponseContentType string `json:"responseContentType"`
	ResponseCode        int    `json:"responseCode"`
	SqliteDatabase      string `json:"sqliteDatabase"`
	SmtpServer          string `json:"smtpServer"`
	SmtpPort            int    `json:"smtpPort"`
	SmtpUser            string `json:"smtpUser"`
	SmtpPassword        string `json:"smtpPassword"`
}

func initConfig() (configuration Configuration, err error) {
	// get configuration from config json
	file, err := os.Open(configFile)
	if err != nil {
		return configuration, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&configuration)
	if err != nil {
		return configuration, err
	}

	return configuration, nil
}

func main() {
	errorLogger := log.New(os.Stderr, "", 0)
	debugLogger := log.New(io.Discard, "", 0)

	// get configuration
	configuration, err := initConfig()
	if err != nil {
		errorLogger.Fatalln(err.Error())
	}

	// evaluate command line flags
	var help bool
	var verbose bool
	flags := flag.NewFlagSet("honeytokenWebServer", flag.ContinueOnError)
	flags.BoolVar(&help, "help", false, "Show this help message")
	flags.BoolVar(&help, "h", false, "")
	flags.BoolVar(&verbose, "v", configuration.VerboseOutput, "Show verbose logging.")
	flags.StringVar(&configuration.InterfaceAndPort, "interfaceAndPort", configuration.InterfaceAndPort, "interface and port e.g. localhost:50000 or :50000 for all interfaces")
	flags.StringVar(&configuration.ResponseFile, "responseFile", configuration.ResponseFile, "the file that will be sent as response to every request")
	flags.StringVar(&configuration.ResponseContentType, "responseContentType", configuration.ResponseContentType, "the Content-Type response header")
	flags.IntVar(&configuration.ResponseCode, "responseCode", configuration.ResponseCode, "the response code (e.g. 200)")
	err = flags.Parse(os.Args[1:])
	switch err {
	case flag.ErrHelp:
		help = true
	case nil:
	default:
		errorLogger.Fatalf("error parsing flags: %v", err)
	}
	// If the help flag was set, just show the help message and exit.
	if help {
		printHelp(flags)
		os.Exit(0)
	}

	if verbose {
		debugLogger = log.New(os.Stderr, "DEBUG: ", 0)
	}

	// check for mandatory configuration items
	if configuration.InterfaceAndPort == "" {
		errorLogger.Fatalln("InterfaceAndPort not set in config.json")
	}
	if configuration.ResponseFile == "" {
		errorLogger.Fatalln("ResponseFile not set in config.json")
	}
	if configuration.ResponseContentType == "" {
		errorLogger.Fatalln("ResponseContentType not set in config.json")
	}
	if configuration.ResponseCode == 0 {
		errorLogger.Fatalln("ResponseCode not set in config.json")
	}

	// check if response file exists before starting server
	if !fileExists(configuration.ResponseFile) {
		errorLogger.Println("Response file", configuration.ResponseFile, "does not exist or is a directory")
		os.Exit(1)
	}

	// load database
	err = ConnectDatabase(configuration.SqliteDatabase)
	if err != nil {
		errorLogger.Fatalln("Could not load database " + configuration.SqliteDatabase)
	}

	// init server struct
	srv := &http.Server{
		Addr:         configuration.InterfaceAndPort,
		Handler:      &App{configuration, errorLogger, debugLogger},
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  10 * time.Second,
	}

	// subscribe to SIGINT signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGTERM)

	// start listener
	go func() {
		errorLogger.Println("Starting honeytokenWebServer with interface \"" + configuration.InterfaceAndPort + "\" response file \"" + configuration.ResponseFile + "\" and response Content-Type \"" + configuration.ResponseContentType + "\"")
		// service connections
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			// server closed regularly
		} else if err != nil {
			fmt.Printf("error starting server: %s\n", err)
			os.Exit(1)
		}
	}()

	// wait for system signal
	<-stopChan
	errorLogger.Println("\nShutting down server...")
	err = srv.Close()
	if err != nil {
		errorLogger.Printf("Close server failed: %v\n", err)
	}

	errorLogger.Println("Server stopped")
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// printHelp prints command line parameter help
func printHelp(flags *flag.FlagSet) {
	fmt.Fprintf(flags.Output(), "\nUsage of %s:\n", os.Args[0])
	flags.PrintDefaults()
	fmt.Printf(`

To configure honeytokenWebServer you can also use a config.json file. Example:

	{
		"verbose": false,
		"interfaceAndPort": "localhost:20000",
		"responseFile": "response2.txt",
		"responseContentType": "text/xml; charset=UTF-8"
	}
`)
}
