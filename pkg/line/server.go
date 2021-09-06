package line

import (
	"context"
	"errors"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/go-logr/logr"
	"google.golang.org/genproto/googleapis/rpc/code"

	"github.com/jlandowner/goline"
)

// AuthzServer implements AuthorizationServer in github.com/envoyproxy/go-control-plane/envoy/service/auth/v3
type AuthzServer struct {
	Log    logr.Logger
	Client *goline.Client
}

func (s *AuthzServer) Check(ctx context.Context, req *authv3.CheckRequest) (*authv3.CheckResponse, error) {
	res := &authv3.CheckResponse{}

	s.Log.V(1).Info("incoming request",
		"attributes", req.GetAttributes(),
		"request", req.GetAttributes().GetRequest(),
		"http", req.GetAttributes().GetRequest().GetHttp(),
		"headers", req.GetAttributes().GetRequest().GetHttp().GetHeaders())

	ah, ok := req.GetAttributes().GetRequest().GetHttp().GetHeaders()["Authorization"]
	if !ok {
		s.Log.Error(errors.New("invalid header"), "Authorization header not found")
		res.Status.Code = int32(code.Code_UNAUTHENTICATED)
		return res, nil
	}

	token, err := extractToken(ah)
	if err != nil {
		s.Log.Error(err, "invalid bearer token")
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

	// Append headers
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
