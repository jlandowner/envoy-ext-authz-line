package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/jlandowner/envoy-ext-authz-line/pkg/line"
	"github.com/jlandowner/goline"
)

var o = &options{}

type options struct {
	LINEClientID            string
	Port                    int
	TLSPrivateKeyPath       string
	TLSCertPath             string
	Insecure                bool
	GracefulShutdownSeconds int
}

func main() {
	flag.StringVar(&o.LINEClientID, "line-client-id", "", "LINE Client ID")
	flag.IntVar(&o.Port, "port", 9443, "server listenting port")
	flag.StringVar(&o.TLSPrivateKeyPath, "tls-key", "tls.key", "TLS key file path")
	flag.StringVar(&o.TLSCertPath, "tls-cert", "tls.crt", "TLS certificate file path")
	flag.BoolVar(&o.Insecure, "insecure", false, "start http server not https server")
	flag.IntVar(&o.GracefulShutdownSeconds, "graceful-shutdown-seconds", 10, "proxy graceful shutdown seconds")
	flag.Parse()

	var log logr.Logger
	zapLog, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	log = zapr.NewLogger(zapLog)
	printOptions(log)
	if err := validteOptions(o); err != nil {
		log.Error(err, "option is not valid")
		os.Exit(1)
	}

	ctx := setupSignalHandler(log)

	authz := &line.AuthzServer{
		Log:    log,
		Client: goline.NewClient(o.LINEClientID, http.DefaultClient),
	}

	// Register UsersServer to gRPC Server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", o.Port))
	if err != nil {
		log.Error(err, "failed to listen port", "port", o.Port)
		os.Exit(1)
	}
	srv := grpc.NewServer()
	authv3.RegisterAuthorizationServer(srv, authz)

	// Add grpc.reflection.v1alpha.ServerReflection
	reflection.Register(srv)

	go func() {
		<-ctx.Done()
		log.Info("shutdowning...")
		srv.GracefulStop()
	}()

	// Start server
	if err := srv.Serve(listener); err != nil {
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
