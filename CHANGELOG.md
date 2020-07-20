# Honeycomb Kubernetes Agent Changelog

## 1.5.0 2019-10-22

Features

- new [drop_event](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/main/docs/configuration-reference.md#drop_event) and [keep_event](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/main/docs/configuration-reference.md#keep_event) processors, for doing simple filtering on events. [#52]
- new [route_event](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/main/docs/configuration-reference.md#route_event) processor for routing events to different datasets based on content. [#53](https://github.com/honeycombio/honeycomb-kubernetes-agent/pull/53)

Thanks @Spindel for these contributions!

## 1.4.0 2019-10-11

Features

- new [regex](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/main/docs/configuration-reference.md#regex) parser, allowing lines to be processed with a list of [RE2](https://github.com/google/re2/wiki/Syntax) regular expressions. [#51](https://github.com/honeycombio/honeycomb-kubernetes-agent/pull/51)

## 1.3.3 2019-08-23 Update Recommended

Fixes

- Supports new system pod log path pattern used by Kubernetes 1.13 and newer. Previous versions of k8s stored pod logs at `/var/log/pods/<configHash>/*` but newer K8s versions write to `/var/log/pods/<configHash>/<containerName>/*`.

## 1.3.2 2019-06-21 Update recommended

Fixes

- Supports new log path pattern for container filtering used in K8s versions published after [this](https://github.com/kubernetes/kubernetes/pull/74441) change. If using containerNames in your watchers configuration and have stopped receiving log data for those containers, this is likely your issue.

## 1.3.1 2019-06-12 Update recommended

Fixes

- Supports new log path pattern used in K8s versions published after [this](https://github.com/kubernetes/kubernetes/pull/74441) change. If you upgraded your K8s cluster and stopped receiving all log data, this is likely your issue.

## 1.3.0 2019-05-06

Features

- New `additional_field` processor for adding arbitrary fields to events. Docs [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/main/docs/configuration-reference.md#additional_fields).
- New `rename_field` processor for renaming event fields. See docs [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/main/docs/configuration-reference.md#rename_field).
- New global `additionalFields` option for adding arbitrary fields to _all_ events sent by the agent. Click [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/main/docs/configuration-reference.md#additionalfields) for more information.

Improvements

- `timefield` processor can now understand time fields of type `time.Time` in addition to string and int. More context [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/issues/35).

## 1.2.0 2019-05-06

Semver introduced, all changes prior to May 6, 2019 included in 1.2.0.
