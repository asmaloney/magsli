package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
)

const listenAddr = ":443"

// These are our command-line options
var (
	mailGunAPIKey     string
	slackWebhookURL   string
	serverSSLCertFile string
	serverSSLKeyFile  string
	showVersion       bool
)

// These are set by the build command using -ldflags (see Makefile)
var version = "not set"
var githash = "not set"
var buildstamp = "not set"

func init() {
	flag.StringVar(&mailGunAPIKey, "mailgun", os.Getenv("MAILGUN_API_KEY"), "MailGun API Key")
	flag.StringVar(&slackWebhookURL, "slack", os.Getenv("SLACK_WEBHOOK_URL"), "Slack webhook to post to")
	flag.StringVar(&serverSSLCertFile, "sslCert", os.Getenv("SERVER_SSL_CERT_FILE"), "Path to server certificate file")
	flag.StringVar(&serverSSLKeyFile, "sslKey", os.Getenv("SERVER_SSL_KEY_FILE"), "Path to server private key file")

	flag.BoolVar(&showVersion, "v", false, "Show version information and exit")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
}

func main() {
	flag.Parse()

	if showVersion {
		printVersionInfo()
		os.Exit(1)
	}

	if mailGunAPIKey == "" ||
		slackWebhookURL == "" ||
		serverSSLCertFile == "" ||
		serverSSLKeyFile == "" {
		printUsage()
		os.Exit(1)
	}

	http.HandleFunc("/", handler)

	printVersionInfo()
	slog.Info("listening on", "IP", listenAddr)

	err := http.ListenAndServeTLS(listenAddr, serverSSLCertFile, serverSSLKeyFile, nil)

	if err != nil {
		log.Fatal(err)
	}
}

func printVersionInfo() {
	fmt.Printf("magsli %v\n", version)
	fmt.Printf("  git hash: %v\n", githash)
	fmt.Printf("  built (UTC): %v\n", buildstamp)
}

func printUsage() {
	fmt.Printf("Usage: magsli [options]\n")
	fmt.Println("Options:")

	flag.PrintDefaults()
}
