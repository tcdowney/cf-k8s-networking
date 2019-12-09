# example: routing to system components

1. make CA and wildcard cert
    (customize the script with the system domain for your environment)
    ```
    ./generate-tls-certs.sh
    ```

2. store the wildcard cert in kubernetes so it can be served by the istio ingress gateway
    ```
    kubectl create -n istio-system secret tls istio-ingressgateway-certs --cert sys-wildcard.crt --key sys-wildcard.key
    ```

    This command isn't idempotent, and new secrets may not get picked up by running ingress pods.
    So, to *change* the secret, the process is 3 steps:
    ```
    kubectl -n istio-system delete secret istio-ingressgateway-certs
    kubectl -n istio-system create secret tls istio-ingressgateway-certs --cert sys-wildcard.crt --key sys-wildcard.key
    kubectl -n istio-system delete pod -n istio-system -l istio=ingressgateway
    ```


3. install an istio gateway for your system domain
   (customize the yaml with the system domain for your environment)
    ```
    kubectl apply -f example-sys-gateway.yaml
    ```

4. install your system component
   (customize with the component and route)
    ```
    kubectl apply -f example-sys-component.yaml
    ```

5. test it out
    ```
    curl -v --cacert sys-ca.crt https://some-cf-component.sys.eirini-dev-1.routing.lol
    ```