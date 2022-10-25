package main

import (
	// import built-in packages
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
	// import local packages
	"httpbin/app"
	"httpbin/app/utils/info"
)

const (
	defaultPort = 8080
)

var (
	port            int
	maxBodySize     int64
	maxDuration     time.Duration
	useRealHostname bool
)

func main() {
	startTime := time.Now().UnixNano()
	flag.IntVar(&port, "port", defaultPort, "Port to listen on")
	flag.Int64Var(&maxBodySize, "max-body-size", app.DefaultMaxBodySize, "Maximum size of request or response, in bytes")
	flag.DurationVar(&maxDuration, "max-duration", app.DefaultMaxDuration, "Maximum duration a response may take")
	flag.BoolVar(&useRealHostname, "use-real-hostname", false, "Expose value of os.Hostname() in the /hostname endpoint instead of dummy value")
	flag.Parse()

	// Command line flags take precedence over environment vars, so we only
	// check for environment vars if we have default values for our command
	// line flags.
	var err error
	if maxBodySize == app.DefaultMaxBodySize && os.Getenv("MAX_BODY_SIZE") != "" {
		maxBodySize, err = strconv.ParseInt(os.Getenv("MAX_BODY_SIZE"), 10, 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid value %#v for env var MAX_BODY_SIZE: %s\n\n", os.Getenv("MAX_BODY_SIZE"), err)
			flag.Usage()
			os.Exit(1)
		}
	}
	if maxDuration == app.DefaultMaxDuration && os.Getenv("MAX_DURATION") != "" {
		maxDuration, err = time.ParseDuration(os.Getenv("MAX_DURATION"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid value %#v for env var MAX_DURATION: %s\n\n", os.Getenv("MAX_DURATION"), err)
			flag.Usage()
			os.Exit(1)
		}
	}
	if port == defaultPort && os.Getenv("PORT") != "" {
		port, err = strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid value %#v for env var PORT: %s\n\n", os.Getenv("PORT"), err)
			flag.Usage()
			os.Exit(1)
		}
	}

	// useRealHostname will be true if either the `-use-real-hostname`
	// arg is given on the command line or if the USE_REAL_HOSTNAME env var
	// is one of "1" or "true".
	if useRealHostnameEnv := os.Getenv("USE_REAL_HOSTNAME"); useRealHostnameEnv == "1" || useRealHostnameEnv == "true" {
		useRealHostname = true
	}

	logger := log.New(os.Stderr, "", 0)

	// A hacky log helper function to ensure that shutdown messages are
	// formatted the same as other messages.  See StdLogObserver in
	// httpbin/middleware.go for the format we're matching here.
	serverLog := func(msg string, args ...interface{}) {
		const (
			logFmt  = "time=%q msg=%q"
			dateFmt = "2006-01-02T15:04:05.9999"
		)
		logger.Printf(logFmt, time.Now().Format(dateFmt), fmt.Sprintf(msg, args...))
	}

	opts := []app.OptionFunc{
		app.WithMaxBodySize(maxBodySize),
		app.WithMaxDuration(maxDuration),
		app.WithObserver(app.StdLogObserver(logger)),
	}
	if useRealHostname {
		hostname, err := os.Hostname()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: use-real-hostname=true but hostname lookup failed: %s\n", err)
			os.Exit(1)
		}
		opts = append(opts, app.WithHostname(hostname))
	}
	hi := app.New(opts...)

	listenAddr := net.JoinHostPort("0.0.0.0", strconv.Itoa(port))

	server := &http.Server{
		Addr:    listenAddr,
		Handler: hi.Handler(),
	}

	// shutdownCh triggers graceful shutdown on SIGINT or SIGTERM
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	// exitCh will be closed when it is safe to exit, after graceful shutdown
	exitCh := make(chan struct{})

	go func() {
		sig := <-shutdownCh
		serverLog("shutdown started by signal: %s", sig)

		shutdownTimeout := maxDuration + 1*time.Second
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(ctx); err != nil {
			serverLog("shutdown error: %s", err)
		}

		close(exitCh)
	}()

	var listenErr error

	info.ShowInfo(startTime)

	listenErr = server.ListenAndServe()

	if listenErr != nil && listenErr != http.ErrServerClosed {
		logger.Fatalf("failed to listen: %s", listenErr)
	}

	<-exitCh
	serverLog("shutdown finished")
}
