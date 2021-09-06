# Contour exmaple

Example configurations to enable external authorization in [contour](https://projectcontour.io/)

## How to use with contour

0. Preconditions: [contour](https://projectcontour.io/getting-started/) and [cert-manager](https://cert-manager.io/docs/installation/) are installed

1. Deploy authz server.

```sh
kubectl create -f https://raw.githubusercontent.com/jlandowner/envoy-ext-authz-line/main/kubernetes/authz-server-cert-manager.yaml
```

2. Edit LINE Client ID (LINE Channel ID)

Edit [LINE Client ID](https://developers.line.biz/en/reference/line-login) to yours.

```sh
kubectl edit deploy envoy-ext-authz-line -n envoy-ext-authz-line
```

```yaml
...
    containers:
      - image: ghcr.io/jlandowner/envoy-ext-authz-line:latest
        name: envoy-ext-authz-line
        args:
        - --line-client-id=YOUR_CLIENT_ID # edit here
        - --port=9443
...
```

3. Apply extention service

```sh
kubectl create -f https://raw.githubusercontent.com/jlandowner/envoy-ext-authz-line/main/kubernetes/contour/extention-service.yaml
```

5. Create your own HTTPProxy with authorization

Contour [HTTPProxy](https://projectcontour.io/docs/latest/config/fundamentals/) is an alternative to Kubernetes Ingress.

Bring your own HTTPProxy and update to use authorization.

See [httpproxy.yaml](https://github.com/jlandowner/envoy-ext-authz-line/blob/main/kubernetes/contour/httpproxy.yaml) as example. 

## Reference

https://projectcontour.io/guides/external-authorization/

https://projectcontour.io/docs/latest/config/api/#projectcontour.io/v1.ExtensionService
