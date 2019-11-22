# Agent Configuration Reference

The agent's behavior is determined by a configuration file that uses a simple YAML syntax. Ordinarily, you'll create this file as a Kubernetes `ConfigMap` that will be mounted inside the agent container.

## Basic Configuration
The only required part of the configuration file is a list of `watchers`:
```
---
watchers:
- labelSelector: "app=nginx"
  parser: nginx
  dataset: kubernetes-nginx

- labelSelector: "app=frontend"
  parser: json
  dataset: kubernetes-frontend
```

Each block in the `watchers` list describes a set of pods whose logs you want
to handle in a specific way, and has the following keys:

key | required? | type | description
:--|:--|:--|:--
labelSelector | yes* | string | A Kubernetes [label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors) identifying the set of pods to watch.
parser | yes | string | Describes how this watcher should parse events.
dataset | yes | string | The dataset that this watcher should send events to.
containerName | no | string | If you only want to consume logs from one container in a multi-container pod, the name of the container to watch.
processors | no | list | A list of [processors](#processors) to apply to events after they're parsed

### Validating a configuration file
To check a configuration file without needing to deploy it into the cluster,
you can run the agent container locally with the `--validate` flag:
```
docker run -v /FULL/PATH/TO/YOUR/config.yaml:/etc/honeycomb/config.yaml honeycombio/honeycomb-kubernetes-agent:head --validate
```

## Parsers
Currently, the following parsers are supported:

### json
Parses logs in JSON format

### regex

Parses logs using a list of regex expressions. Uses the [RE2 Syntax](https://github.com/google/re2/wiki/Syntax). For each line processed, the first regex in the list to match is used. Note that adding many regex expressions can hurt performance.

```yaml
parser:
  name: regex
  options:
    # lines contain either tree data or city/state data
    expressions:
      - "(?P<species>[A-z]+) (?P<height>[0-9]{2}[0-9]?)"
      - "(?P<city>[A-z ]+),(?P<state>[A-z]{2})"
```

### nginx
Parses NGINX access logs.

If you're using a custom NGINX log format, you can specify the format using the following configuration:
```
parser:
  name: nginx
  options:
    log_format: '$remote_addr - $remote_user [$time_local] "$request" $status ...'
```

### glog
Parses logs produced by [glog](https://github.com/golang/glog), which look like this:
```
I0719 23:09:54.422170       1 kube.go:118] Node controller sync successful
```

This format is commonly used by Kubernetes system components such as the API server.

### redis
Parses logs produced by [redis](https://redis.io) 3.0+, which look like this:
```
1:M 08 Aug 22:59:58.739 * Background saving started by pid 43
```

### keyval
Parses logs in `key=value` format.

Key-value formatted logs often have a special prefix, such as a timestamp. You
can specify a regular expression to parse that prefix using the following
configuration:
```
parser:
  name: keyval
  options:
    prefixRegex: "(?P<timestamp>[0-9:\\-\\.TZ]+) AUDIT: "
```

### audit
Parses [Kubernetes audit logs](https://kubernetes.io/docs/tasks/debug-application-cluster/audit/#audit-logs).

### nop
Does no parsing on logs, and returns an event with the entire contents of the log line in a `"log"` field.

More parsers will be added in the future. If you'd like to see support for additional log formats, please open an issue or email support@honeycomb.io!

## Processors
Processors transform events after they're parsed. Currently, the following processors are supported:

### additional_fields

The `additional_fields` processor accepts a static map of field names and values and appends those to every event it processes. These values will overwrite existing fields of the same name, if they exist:

**Example:**

```yaml
processors:
  - additional_fields:
      environment: production
      owner: me@example.com
```

### sample
The `sample` processor will only send a subset of events to Honeycomb. Honeycomb natively supports sampled event streams, allowing you to send a representative subset of events while still getting high-fidelity query results.

**Options:**

key | type | description
:--|:--|:--
type | `"static"` or `"dynamic"` | How events should be sampled.
rate | integer | The rate at which to sample events. Specifying a sample rate of 20 will cause one in 20 events to be sent.
keys | list of strings | The list of field keys to use when doing dynamic sampling.
windowSize | int | How often to refresh estimated sample rates when doing dynamic sampling, in seconds. Defaults to 30 seconds.

### drop_field

The `drop_field` processor will remove the specified field from all events before sending them to Honeycomb. This is useful for removing sensitive information from events.

**Options:**

key | value | description
:--|:--|:--
field | string | The name of the field to drop.

### drop_event

The `drop_event` processor will remove all events where the specified field
exactly matches one of the values in the blacklist.

This can be used to filter datasets from _Ingress_ to _Service_ based on which
_Service_ or _Namespace_ is used.

Events that do not have a `field` matching this configuration will be kept.

**Options:**

key | value | description
:--|:--|:--
field | string | The name of the event field to match against the blacklist
values | list  | The set of field values that cause this processor to drop an event

### keep_event

The `keep_event` processor will remove all events NOT matching one of the
whitelisted values. Careless configuration of this filter will drop all events.

It is effectively the inverse of `drop_event`.

Events that do not have a `field` matching the configuration will be kept to
avoid accidental data loss.

**Options:**
key | value | description
:--|:--|:--
field | string | The name of the field to match against the whitelist
values | list  | The name of the event field to match against the whitelist


### route_event

The `route_event` processor will route events to a different dataset depending
on the value of a field.  This is mostly useful for load balancing web servers
or other resources that serve multiple different backends, but where you still
wish to track a common identifier between systems.


**Options:**
key | value | description
:--|:--|:--
field | string | The name of the event field to base routing on
routes | list | A list of routing configurations

**Routes:**

key | value | description
:--|:--|:--
value   | string | Route if the event field is exactly this value
dataset | string | The dataset to route to


Example configuration
```
processors:
- route_events:
    field: host
    routes:
      - value: api.example.com
        dataset: api
      - value: www.example.com
        dataset: web
```

The above configuration would route events where `host` exists and is equal to
`api.example.com` to the `api` dataset. It would also route requests to
`www.example.com` to the `www` dataset. All others would be routed to the
default dataset for this watcher.

### rename_field

The `rename_field` processor will rename the specified field in all events, if it exists, before sending them to Honeycomb. You can use this to standardize field names across data sources, for example. Note that if the `new` field already exists, it will be overwritten with the value in the `original` field.

**Options:**

key | type | description
:--|:--|:--
original | string | Name of field to be renamed. Required.
new | string | The new field name to use. Required.


### timefield
The `timefield` processor will replace the default timestamp in an event with
one extracted from a specific field in the event.

**Options:

key | value | description
:--|:--|:--
field | string | The name of the field containing the timestamp
format | string | The format of the timestamp found in timefield, in strftime or [Golang](https://golang.org/pkg/time/#pkg-constants) format

_Note_: This processor isn't generally necessary when collecting pod logs. The
agent will automatically use the timestamp recorded by the Docker json-log
driver. It's useful when parsing logs that live at a particular path on the
node filesystem, such as Kubernetes audit logs.

### request_shape

The `request_shape` processor will take a field representing an HTTP request, such as `GET /api/v1/users?id=22 HTTP/1.1`, and unpack it into its constituent parts.

**Options:**

key | value | description
:--|:--|:--
field | string | The name of the field containing the HTTP request (e.g., `"request"`)
patterns | list of strings | A list of URL patterns to match when unpacking the request
queryKeys | list of strings | A whitelist of keys in the URL query string to unpack
prefix | string | A prefix to prepend to the unpacked field names

For example, with the following configuration:

```
processors:
- request_shape:
    field: request
    patterns:
    - /api/:version/:resource
    queryKeys:
    - id
```

the request_shape processor will expand the event

```
{"request": "GET /api/v1/users?id=22 HTTP/1.1", ...}
```

into

```
{
    "request": "GET /api/v1/users?id=22 HTTP/1.1",
    "request_method": "GET",
    "request_protocol_version": "HTTP/1.1",
    "request_uri": "/api/v1/users?id=22",
    "request_path": "/api/v1/users",
    "request_query": "id=22",
    "request_shape": "/api/:version/:resource?id=?",
    "request_path_version": "v1",
    "request_path_resource": "users",
    "request_pathshape": "/api/:version/:resource",
    "request_queryshape": "id=?",
    "request_query_id": "22",
    ...
}
```

## Global Configuration Options

### additionalFields

You can append arbitrary key/value pairs to all events processed by the agent by defining the `additionalFields` section of the config. These values will overwrite existing fields of the same name, if they exist:

```yaml
writekey: "YOUR_HONEYCOMB_WRITEKEY_HERE"
watchers:
  - # ...
additionalFields:
  cluster_name: my cluster
  owner: me@example.com
```

## Sample configurations

Here are some example configurations for the Honeycomb agent.

 * Parse logs from pods labelled with app: nginx:

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

 * Send logs from different services to different datasets:

    ```
    ---
    writekey: "YOUR_HONEYCOMB_WRITEKEY_HERE"
    watchers:
      - labelSelector: "app=nginx"
        parser: nginx
        dataset: nginx-kubernetes

      - labelSelector: "app=frontend-web"
        parser: json
        dataset: frontend
    ```

 * Sample events from a frontend-web deployment: only send one in 20 events from the prod namespace, and one in 10 events from the staging namespace.

    ```
    ---
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

 * Only process logs from the sidecar container in a multi-container pod:

    ```
    ---
    writekey: "YOUR_HONEYCOMB_WRITEKEY_HERE"
    watchers:
      - labelSelector: "app=frontend-web"
        containerName: sidecar
        parser: json
        dataset: frontend
    ```
