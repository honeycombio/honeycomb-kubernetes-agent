Here are some example configurations for the Honeycomb agent:


Parse logs from pods labelled with `app: nginx`:
```
---
writekey: "YOUR_HONEYCOMB_WRITEKEY_HERE"
watchers:
  - labelSelector: app=nginx
    parser: nginx
    dataset: nginx-kubernetes

    processors:
    - request_shape:
        field: request
```

Send logs from different services to different datasets:
```
writekey: "YOUR_HONEYCOMB_WRITEKEY_HERE"
watchers:
  - labelSelector: "app=nginx"
    parser: nginx
    dataset: nginx-kubernetes

  - labelSelector: "app=frontend-web"
    parser: json
    dataset: frontend
```


Sample events from a `frontend-web` deployment: only send one in 20 events from
the `prod` namespace, and one in 10 events from the `staging` namespace.
```
writekey: "YOUR_HONEYCOMB_WRITEKEY_HERE"
watchers:
  - labelSelector: "app=frontend-web"
    namespace: prod
    parser: json
    dataset: frontend

    processors:
      - sample:
          type: static
          rate: 20
      - drop_field:
        field: user_email

  - labelSelector: "app=frontend-web"
    namespace: staging
    parser: json
    dataset: frontend

    processors:
      - sample:
          type: static
          rate: 10
```

Only process logs from the `sidecar` container in a multi-container pod:
```
---
writekey: "YOUR_HONEYCOMB_WRITEKEY_HERE"
watchers:
  - labelSelector: "app=frontend-web"
    containerName: sidecar
    parser: json
    dataset: frontend
```
