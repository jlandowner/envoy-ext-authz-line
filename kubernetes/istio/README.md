# Istio example

Example configurations to enable external authorization in [istio](https://istio.io/)

## How to use with istio

0. Preconditions: [Istio](https://istio.io/latest/docs/setup/getting-started/) is installed

1. Deploy authz server.

```sh
kubectl create -f https://raw.githubusercontent.com/jlandowner/envoy-ext-authz-line/kubernetes/authz-server.yaml
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

3. Update istio global mesh config

Update istio global mesh config to add [extensionProviders](https://istio.io/latest/docs/reference/config/istio.mesh.v1alpha1/#MeshConfig-ExtensionProvider)

```sh
kubectl edit configmap istio -n istio-system
```

```yaml
data:
  mesh: |-
    # Add the following content to define the external authorizers.
    extensionProviders:
    - name: "envoy-ext-authz-line"
      envoyExtAuthzGrpc:
        service: "envoy-ext-authz-line.envoy-ext-authz-line.svc.cluster.local"
        port: "9443"
```

Restart istiod

```sh
kubectl rollout restart deployment/istiod -n istio-system
```

4. Apply AuthorizationPolicy to your app

Apply [AuthorizationPolicy](https://istio.io/latest/docs/reference/config/security/authorization-policy/) in your app namespace.

```sh
APP_NAMESPACE=YOUR_APP_NAMESPACE
kubectl create -n $APP_NAMESPACE -f https://raw.githubusercontent.com/jlandowner/envoy-ext-authz-line/kubernetes/istio/authz-policy.yaml
```

Edit with your app info

```sh
kubectl edit AuthorizationPolicy envoy-ext-authz-line
```

```yaml
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: envoy-ext-authz-line
spec:
  selector:
    matchLabels:
      app: example-service # edit here with your app labels that you want to make secure.
  action: CUSTOM
  provider:
    name: envoy-ext-authz-line
  rules:
    - {}
```

That's all.

## Reference

https://istio.io/latest/docs/tasks/security/authorization/authz-custom/#enable-with-external-authorization

https://istio.io/latest/docs/reference/config/security/authorization-policy/
