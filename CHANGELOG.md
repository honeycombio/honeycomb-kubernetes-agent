# Honeycomb Kubernetes Agent Changelog

## 2.3.1 2022-01-21

### Fixed

- fix nil pointer deref when includeNodeLabels is enabled (#218) | [@asdvalenzuela](https://github.com/asdvalenzuela)

## 2.3.0 2022-01-12

### Added

- add node metadata to metrics events (#200) | [@asdvalenzuela](https://github.com/asdvalenzuela)

### Fixed

- fix: check for existence of optional CPU stats to prevent exceptions (#206) | [@JamieDanielson](https://github.com/JamieDanielson)

### Maintenance

- gh: add re-triage workflow (#203) | [@vreynolds](https://github.com/vreynolds)
- maint: update releasing notes for ci workflow and helm chart update (#205) | [JamieDanielson](https://github.com/JamieDanielson)
- Bump github.com/mitchellh/mapstructure from 1.4.2 to 1.4.3 (#208) | [dependabot](https://github.com/dependabot)
- Update dependabot label (#210) | [@vreynolds](https://github.com/vreynolds)
- switch to supported versions of k8s libraries (#211) | [@vreynolds](https://github.com/vreynolds)
- Bump github.com/honeycombio/libhoney-go from 1.15.6 to 1.15.7 (#212) | [dependabot](https://github.com/dependabot)
- test script needs directory (#214) | [JamieDanielson](https://github.com/JamieDanielson)

## 2.2.1 2021-12-23

### Fixes

- container metadata missing (#198) | [@vreynolds](https://github.com/vreynolds)
- Fix base selector for logs (#183) | [@puckpuck](https://github.com/puckpuck)

### Maintenance

- docs: update developing for new build (#199) | [@vreynolds](https://github.com/vreynolds)
- Makes image non transparent (#191) | [@bdarfler](https://github.com/bdarfler)
- add release process (#190) | [@puckpuck](https://github.com/puckpuck)
- Update dependabot.yml (#195) | [@vreynolds](https://github.com/vreynolds)
- maint: Update ownership and community health files (#194) | [@JamieDanielson](https://github.com/JamieDanielson)
- update labeler to trigger with power (#187) | [Robb Kidd]
- Bump k8s.io/client-go from 0.22.3 to 0.22.4 (#193) | [dependabot[bot]]
- Bump k8s.io/api from 0.22.3 to 0.22.4 (#192) | [dependabot[bot]]
- Bump github.com/honeycombio/honeytail from 1.5.0 to 1.6.0 (#188) | [dependabot[bot]]
- Bump github.com/honeycombio/libhoney-go from 1.15.5 to 1.15.6 (#189) | [dependabot[bot]]
- Bump k8s.io/client-go from 0.22.2 to 0.22.3 (#186) | [dependabot[bot]]
- Bump github.com/honeycombio/libhoney-go from 1.15.4 to 1.15.5 (#184) | [dependabot[bot]]
- Bump k8s.io/client-go from 0.21.3 to 0.22.2 (#180) | [dependabot[bot]]
- Bump github.com/mitchellh/mapstructure from 1.4.1 to 1.4.2 (#179) | [dependabot[bot]]

## 2.2.0 2021-09-07

- Adds multiple architecture image support (amd64, arm64) [#164](https://github.com/honeycombio/honeycomb-kubernetes-agent/pull/164)

## 2.1.3 2021-03-22

- Fixes [#144](https://github.com/honeycombio/honeycomb-kubernetes-agent/issues/144) leveraging a new `NODE_IP` environment variable

## 2.1.2 2021-03-03

- Update Go to 1.15.8, fixing http2 race condition [#143](https://github.com/honeycombio/honeycomb-kubernetes-agent/pull/143)

## 2.1.1 2021-01-05

- Drastically reduces image size
- Fixes panic in K8s 1.20

## 2.1.0 2020-11-13

- Introduces native metrics collection support for nodes, pods, containers, and volumes.
- Includes status collection for pods and containers.

## 2.0.0 2020-05-13

- Compatible with latest Kubernetes release (1.18)
- support for CRI runtime
- support for kind (Kubernetes IN Docker)

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
