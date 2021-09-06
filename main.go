package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"

	"github.com/jlandowner/envoy-ext-authz-line/pkg/line"
	"github.com/jlandowner/goline"
)

var o = &options{}

type options struct {
	LINEClientID      string
	Port              int
	TLSPrivateKeyPath string
	TLSCertPath       string
	Insecure          bool
	Debug             bool
}

func main() {
	flag.StringVar(&o.LINEClientID, "line-client-id", "", "LINE Client ID")
	flag.IntVar(&o.Port, "port", 9443, "server listenting port")
	flag.StringVar(&o.TLSPrivateKeyPath, "tls-key", "tls.key", "TLS key file path")
	flag.StringVar(&o.TLSCertPath, "tls-cert", "tls.crt", "TLS certificate file path")
	flag.BoolVar(&o.Insecure, "insecure", false, "start http server not https server")
	flag.BoolVar(&o.Debug, "debug", false, "whether zap.NewProduction or zap.NewDevelopment")
	flag.Parse()

	var log logr.Logger
	var zapLog *zap.Logger
	if o.Debug {
		l, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		zapLog = l
	} else {
		l, err := zap.NewProduction()
		if err != nil {
			panic(err)
		}
		zapLog = l
	}

	log = zapr.NewLogger(zapLog)
	printOptions(log)
	if err := validteOptions(o); err != nil {
		log.Error(err, "option is not valid")
		os.Exit(1)
	}

	ctx := setupSignalHandler(log)

	authz := &line.AuthzServer{
		Log:    log.WithName("AuthzServer"),
		Client: goline.NewClient(o.LINEClientID, http.DefaultClient),
	}

	// Register UsersServer to gRPC Server
	lis, err := setupListener(o)
	if err != nil {
		log.Error(err, "failed to setup listener", "port", o.Port)
		os.Exit(1)
	}

	// Start server
	if err := authz.Run(ctx, lis); err != nil {
		log.Error(err, "failed to start server")
	}
}

func printOptions(log logr.Logger) {
	rv := reflect.ValueOf(*o)
	rt := rv.Type()
	options := make([]interface{}, rt.NumField()*2)

	for i := 0; i < rt.NumField(); i++ {
		options[i*2] = rt.Field(i).Name
		options[i*2+1] = rv.Field(i).Interface()
	}
	log.Info("options", options...)
}

func validteOptions(o *options) error {
	if o.LINEClientID == "" {
		return errors.New("LINEClientID is required")
	}
	return nil
}

func setupListener(o *options) (net.Listener, error) {
	if o.Insecure {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", o.Port))
		if err != nil {
			return nil, fmt.Errorf("failed to listen port: %w", err)
		}
		return ln, nil

	} else {
		cer, err := tls.LoadX509KeyPair(o.TLSCertPath, o.TLSPrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load keypair: %w", err)
		}
		cfg := tls.Config{Certificates: []tls.Certificate{cer}}
		ln, err := tls.Listen("tcp", fmt.Sprintf(":%d", o.Port), &cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to listen port: %w", err)
		}
		return ln, nil
	}
}

func setupSignalHandler(log logr.Logger) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.WithName("signalHandler").Info("got shutdown signals. do gracefull shutdown")
		cancel()
		<-c
		log.WithName("signalHandler").Info("got shutdown signals again. force quit")
		os.Exit(1) // second signal. Exit directly.
	}()

	return ctx
}
