# Honeycomb Kubernetes Agent Changelog

## 1.3.1 2019-06-12 Update recommended

Fixes

- Supports new log path pattern used in K8s versions published after [this](https://github.com/kubernetes/kubernetes/pull/74441) change. If you upgraded your K8s cluster and stopped receiving all log data, this is likely your issue.

## 1.3.0 2019-05-06

Features

- New `additional_field` processor for adding arbitrary fields to events. Docs [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/master/docs/configuration-reference.md#additional_fields).
- New `rename_field` processor for renaming event fields. See docs [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/master/docs/configuration-reference.md#rename_field).
- New global `additionalFields` opton for adding arbitrary fields to *all* event sent by the agent. Click [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/blob/master/docs/configuration-reference.md#additionalfields) for more information.

Improvements

- `timefield` processor can now understand time fields of type `time.Time` in addition to string and int. More context [here](https://github.com/honeycombio/honeycomb-kubernetes-agent/issues/35).

## 1.2.0 2019-05-06

Semver introduced, all changes prior to May 6, 2019 included in 1.2.0.
