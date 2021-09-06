package authz

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"

	authv3 "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/go-logr/zapr"
	"github.com/jlandowner/goline"
	"go.uber.org/zap"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"

	"github.com/stretchr/testify/assert"
)

// TODO: in order to test OK case, replace VALID_CLIENT_ID to the valid value.
var clientid = "VALID_CLIENT_ID"

func TestLINEAuthzServer(t *testing.T) {
	ctx := context.Background()
	assert := assert.New(t)

	zapLog, err := zap.NewDevelopment()
	assert.Nil(err)

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:0"))
	assert.Nil(err)
	assert.NotNil(lis)

	addr, ok := lis.Addr().(*net.TCPAddr)
	assert.True(ok)

	serv := &LINEAuthzServer{
		Log:    zapr.NewLogger(zapLog),
		Client: goline.NewClient(clientid, http.DefaultClient),
	}
	// Start the test server on random port.
	go serv.Run(ctx, lis)

	// Prepare the gRPC request.
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", addr.Port), grpc.WithInsecure())
	assert.Nil(err)
	assert.NotNil(conn)

	defer conn.Close()
	grpcV3Client := authv3.NewAuthorizationClient(conn)

	cases := []struct {
		name   string
		header string
		want   int
	}{
		// TODO: in order to test OK case, uncomment below and replace VALID_ACCESS_TOKEN to the valid value.
		// {
		// 	name:   "bearer",
		// 	header: "Bearer VALID_ACCESS_TOKEN",
		// 	want:   int(code.Code_OK),
		// },
		{
			name:   "invalid bearer",
			header: "Bearer xxx",
			want:   int(code.Code_UNAUTHENTICATED),
		},
		{
			name:   "not bearer",
			header: "deny",
			want:   int(code.Code_UNAUTHENTICATED),
		},
		{
			name:   "not auth header",
			header: "",
			want:   int(code.Code_UNAUTHENTICATED),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := &authv3.CheckRequest{
				Attributes: &authv3.AttributeContext{
					Request: &authv3.AttributeContext_Request{
						Http: &authv3.AttributeContext_HttpRequest{
							Host: "localhost",
							Path: "/",
						},
					},
				},
			}
			if tc.header != "" {
				req.Attributes.Request.Http.Headers = map[string]string{authHeader: tc.header}
			}

			resp, err := grpcV3Client.Check(ctx, req)
			assert.Nil(err)

			got := int(resp.Status.Code)
			if got != tc.want {
				t.Errorf("want %d but got %d", tc.want, got)
			}
		})
	}
}
