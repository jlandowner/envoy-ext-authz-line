package authz

import (
	"context"
	"errors"
	"net"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/go-logr/logr"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/jlandowner/goline"
)

const (
	authHeader string = "authorization"
)

// LINEAuthzServer implements the envoy AuthorizationServer
// https://pkg.go.dev/github.com/envoyproxy/go-control-plane/envoy/service/auth/v3#AuthorizationServer
type LINEAuthzServer struct {
	Log    logr.Logger
	Client *goline.Client
}

// Check checks the authorization header. Extract bearer LINE access token from the header and authorize it upstream LINE Login service.
func (s *LINEAuthzServer) Check(ctx context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	res := &authv3.CheckResponse{
		Status: &status.Status{},
	}

	s.Log.V(1).Info("incoming request", "attributes", req.GetAttributes())

	ah, ok := req.GetAttributes().GetRequest().GetHttp().GetHeaders()[authHeader]
	if !ok {
		s.Log.Error(errors.New("authorization header not found"), "invalid header")
		res.Status.Code = int32(code.Code_UNAUTHENTICATED)
		return res, nil
	}

	token, err := extractToken(ah)
	if err != nil {
		s.Log.Error(err, "invalid header")
		res.Status.Code = int32(code.Code_UNAUTHENTICATED)
		return res, nil
	}

	if _, err := s.Client.VerifyAccessToken(ctx, token); err != nil {
		s.Log.Error(err, "failed to verify access token")
		res.Status.Code = int32(code.Code_UNAUTHENTICATED)
		return res, nil
	}

	p, err := s.Client.GetProfile(ctx, token)
	if err != nil {
		s.Log.Error(err, "failed to get profile")
		res.Status.Code = int32(code.Code_UNAUTHENTICATED)
		return res, nil
	}

	// Append user profiles in headers
	res.HttpResponse = &authv3.CheckResponse_OkResponse{
		OkResponse: &authv3.OkHttpResponse{
			Headers: []*corev3.HeaderValueOption{
				{
					Header: &corev3.HeaderValue{
						Key:   goline.HeaderKeyLINEUserID,
						Value: p.UserID,
					},
				},
				{
					Header: &corev3.HeaderValue{
						Key:   goline.HeaderKeyLINEDisplayName,
						Value: p.DisplayName,
					},
				},
				{
					Header: &corev3.HeaderValue{
						Key:   goline.HeaderKeyLINEPictureURL,
						Value: p.PictureURL,
					},
				},
				{
					Header: &corev3.HeaderValue{
						Key:   goline.HeaderKeyLINEStatusMessage,
						Value: p.StatusMessage,
					},
				},
			},
		},
	}

	res.Status.Code = int32(code.Code_OK)
	return res, nil
}

func extractToken(authHeader string) (string, error) {
	arr := strings.Split(authHeader, "Bearer ")
	if len(arr) != 2 {
		return "", errors.New("not bearer")
	}
	return arr[1], nil
}

// Run start grpc server
func (s *LINEAuthzServer) Run(ctx context.Context, lis net.Listener) error {
	if s.Client == nil {
		panic("goline client is nil")
	}
	if lis == nil {
		panic("listener is nil")
	}
	addr, ok := lis.Addr().(*net.TCPAddr)
	if !ok {
		panic("not tcp listener")
	}

	srv := grpc.NewServer()
	authv3.RegisterAuthorizationServer(srv, s)

	// Add grpc.reflection.v1alpha.ServerReflection
	reflection.Register(srv)

	go func() {
		<-ctx.Done()
		s.Log.Info("shutdown grpc server...")
		srv.GracefulStop()
	}()

	// Start server
	s.Log.Info("starting grpc server", "listenPort", addr.Port)
	return srv.Serve(lis)
}
