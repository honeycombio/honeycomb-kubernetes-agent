# Honeycomb Kubernetes Agent Changelog

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

- New `additional_field` processor for adding arbitrary fields to events. Docs [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/master/docs/configuration-reference.md#additional_fields).
- New `rename_field` processor for renaming event fields. See docs [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/master/docs/configuration-reference.md#rename_field).
- New global `additionalFields` option for adding arbitrary fields to _all_ events sent by the agent. Click [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/master/docs/configuration-reference.md#additionalfields) for more information.

Improvements

- `timefield` processor can now understand time fields of type `time.Time` in addition to string and int. More context [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/issues/35).

## 1.2.0 2019-05-06

Semver introduced, all changes prior to May 6, 2019 included in 1.2.0.
