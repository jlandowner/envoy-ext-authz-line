[![Go Report Card](https://goreportcard.com/badge/github.com/jlandowner/envoy-ext-authz-line)](https://goreportcard.com/report/github.com/jlandowner/envoy-ext-authz-line)
[![Go Reference](https://pkg.go.dev/badge/github.com/jlandowner/envoy-ext-authz-line.svg)](https://pkg.go.dev/github.com/jlandowner/envoy-ext-authz-line)
[![GHCR](https://img.shields.io/badge/download-github_packages-brightgreen)](https://github.com/jlandowner/kubernetes-route53-sync/pkgs/container/kubernetes-route53-sync)

# Simple LINE Login implementation for Envoy External Authorization

LINE Login authorization service implementing Envoy External Authorization API.

https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/auth/v3/external_auth.proto

# How it works

[Envoy External Authorization](https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_authz/v3/ext_authz.proto.html) is a feature to intercept user request with custom authorization. 

This package interact with envoy and check the http request header.

Extract LINE access token from header `Authorization: Bearer LINE_ACCESS_TOKEN` and authorize it upstream LINE Login service.

Then if the upstream authorization is passed, append LINE User info set on the header to be able to use at backend service.

See the LINE official docs how to properly manage access token between client and server.

https://developers.line.biz/en/docs/line-login/secure-login-process/

Detail of LINE Login API 

https://developers.line.biz/en/docs/line-login/

# How to use

- [Use with Istio](https://github.com/jlandowner/envoy-ext-authz-line/blob/main/kubernetes/istio/)

- [Use with Contour](https://github.com/jlandowner/envoy-ext-authz-line/blob/main/kubernetes/contour/)

# License
MIT
