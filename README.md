[![Go Report Card](https://goreportcard.com/badge/github.com/jlandowner/envoy-ext-authz-line)](https://goreportcard.com/report/github.com/jlandowner/envoy-ext-authz-line)
[![Go Reference](https://pkg.go.dev/badge/github.com/jlandowner/envoy-ext-authz-line.svg)](https://pkg.go.dev/github.com/jlandowner/envoy-ext-authz-line)
[![GHCR](https://img.shields.io/badge/download-github_packages-brightgreen)](https://github.com/jlandowner/envoy-ext-authz-line/pkgs/container/envoy-ext-authz-line)

# LINE Login authorizer for Envoy External Authorization

Simple LINE Login authorization implementation of Envoy External Authorization protocol.

https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto

# How to use

- [Use with Istio](https://github.com/jlandowner/envoy-ext-authz-line/blob/main/kubernetes/istio/)

- [Use with Contour](https://github.com/jlandowner/envoy-ext-authz-line/blob/main/kubernetes/contour/)

# How it works

[Envoy External Authorization](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto.html) is a feature to intercept user request with custom authorizations. 

This package interact with envoy and check the http request headers.

Extract LINE access token from header `Authorization: Bearer LINE_ACCESS_TOKEN` and authorize it upstream LINE Login service.

See the LINE official docs how to manage access token properly between client and server.

https://developers.line.biz/en/docs/line-login/secure-login-process/

Detail of LINE Login API 

https://developers.line.biz/en/docs/line-login/

## LINE User profile header keys
If the upstream authorization is passed, LINE User profile info will be added to the request headers so that it can be used by the backend services.

The authorized LINE User profile info can be available by the following header keys.

- `LINEUserID`
- `LINEDisplayName`
- `LINEPictureURL`
- `LINEStatusMessage`

See the docs about details of each contents.

https://developers.line.biz/en/reference/line-login/#get-user-profile

# License
MIT
