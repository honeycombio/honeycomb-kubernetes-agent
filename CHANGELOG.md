# Honeycomb Kubernetes Agent Changelog

## 2.7.2 2024-02-06

### Maintenance

- chore(github): set dependabot reviewers correctly (#404) | @lizthegrey
- bump deps (#383) | @TylerHelmuth
- maint: update codeowners to pipeline-team (#396) | @JamieDanielson
- maint: update project workflow for pipeline (#395) | @JamieDanielson
- maint: update codeowners to pipeline (#394) | @JamieDanielson
- maint: update dependabot.yml (#364) | @vreynolds
- maint(deps): bump k8s.io/kubelet from 0.28.2 to 0.29.0 (#397) | @Dependabot
- maint(deps): bump k8s.io/kubelet from 0.29.0 to 0.29.1 (#403) | @Dependabot
- maint(deps): bump github.com/honeycombio/honeytail from 1.8.3 to 1.9.0 (#401) | @Dependabot
- maint(deps): bump github.com/honeycombio/dynsampler-go (#399) | @Dependabot
- maint(deps): bump k8s.io/client-go from 0.29.0 to 0.29.1 (#400) | @Dependabot
- maint(deps): bump github.com/bmatcuk/doublestar/v4 from 4.6.0 to 4.6.1 (#392) | @Dependabot
- maint(deps): bump k8s.io/client-go from 0.28.2 to 0.28.3 (#389) | @Dependabot
- maint(deps): bump k8s.io/api from 0.28.2 to 0.28.3 (#388) | @Dependabot
- maint(deps): bump golang.org/x/net from 0.13.0 to 0.17.0 (#387) | @Dependabot
- maint(deps): bump k8s.io/kubelet from 0.28.1 to 0.28.2 (#384) | @Dependabot
- maint(deps): bump github.com/honeycombio/dynsampler-go (#371) | @Dependabot
- maint(deps): bump k8s.io/client-go from 0.27.2 to 0.27.3 (#374) | @Dependabot
- maint(deps): bump github.com/honeycombio/libhoney-go (#373) | @Dependabot
- maint(deps): bump k8s.io/kubelet from 0.27.2 to 0.27.3 (#372) | @Dependabot
- maint(deps): bump github.com/sirupsen/logrus from 1.9.2 to 1.9.3 (#375) | @Dependabot
- maint(deps): bump github.com/sirupsen/logrus from 1.9.0 to 1.9.2 (#370) | @Dependabot
- maint(deps): bump k8s.io/client-go from 0.27.1 to 0.27.2 (#368) | @Dependabot
- maint(deps): bump k8s.io/kubelet from 0.27.1 to 0.27.2 (#367) | @Dependabot
- maint(deps): bump github.com/stretchr/testify from 1.8.2 to 1.8.4 (#369) | @Dependabot
- maint(deps): bump k8s.io/api from 0.27.1 to 0.27.2 (#366) | @Dependabot
- maint(deps): bump github.com/honeycombio/honeytail from 1.8.2 to 1.8.3 (#363) | @Dependabot
- maint(deps): bump k8s.io/kubelet from 0.26.3 to 0.27.1 (#365) | @Dependabot
- maint(deps): bump k8s.io/client-go from 0.26.3 to 0.27.1 (#362) | @Dependabot

## 2.7.1 2023-04-13

### Maintenance

- maint(deps): bump github.com/honeycombio/dynsampler-go (#356) | [dependabot[bot]](https://github.com/dependabot[bot])
- maint(deps): bump k8s.io/kubelet from 0.26.2 to 0.26.3 (#357) | [dependabot[bot]](https://github.com/dependabot[bot])
- maint(deps): bump k8s.io/client-go from 0.26.2 to 0.26.3 (#355) | [dependabot[bot]](https://github.com/dependabot[bot])
- Update CHANGELOG.md (#350) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- Add labels to build (#354) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- Add LICENSES dir (#353) | [Tyler Helmuth](https://github.com/TylerHelmuth)
- Update workflow to match refinery (#349) | [Tyler Helmuth](https://github.com/TylerHelmuth)

## 2.7.0 2023-03-13

### Fixes
- fix: Validate label selector and paths are mutually exclusive (#334) | @TylerHelmuth
- fix: do not accumulate if pod is not there anymore #333 (#348) | @enc

### Maintenance
- maint(deps): bump k8s.io/kubelet from 0.26.1 to 0.26.2 (#347)
- maint(deps): bump k8s.io/client-go from 0.26.1 to 0.26.2 (#345)
- maint(deps): bump github.com/honeycombio/dynsampler-go (#339)
- maint(deps): bump golang.org/x/net from 0.5.0 to 0.7.0 (#342)
- maint(deps): bump github.com/stretchr/testify from 1.8.1 to 1.8.2 (#344)
- docs: Update keyval docs (#340) | @TylerHelmuth
- docs: Add paths and exclude docs (#338) | @TylerHelmuth
- maint(deps): bump k8s.io/client-go from 0.25.2 to 0.26.1 (#337)
- maint(deps): bump k8s.io/kubelet from 0.25.3 to 0.26.1 (#335)
- maint(deps): bump github.com/bmatcuk/doublestar/v4 from 4.3.0 to 4.6.0 (#326)
- Bump github.com/honeycombio/honeytail from 1.8.1 to 1.8.2 (#316)
- maint: don't spam the logs with filtered-out filenames (#331) | @kentquirk
- Update CODEOWNERS (#328) | @TylerHelmuth
- chore: update workflows (#327) | @kentquirk
- chore: add maint: prefix to dependabot prs (#319) | @JamieDanielson
- ci: update validate PR title workflow (#312) | @pkanal
- ci: validate PR title (#311) | @pkanal

## 2.6.0 2022-11-21

### Added

- Add MinEventsPerSec to config (#308) | [@kentquirk](https://github.com/kentquirk)

## 2.5.6 2022-11-16

### Fixes

- Add support for building sample keys from integer fields (#304) | [@puckpuck](https://github.com/puckpuck)

## 2.5.5 2022-11-7

### Fixes

- Put the k8s metadata processor first (#302) | [@kentquirk](https://github.com/kentquirk)

### Maintenance

- Bump github.com/honeycombio/libhoney-go from 1.17.0 to 1.18.0 (#300)
- Bump k8s.io/kubelet from 0.25.2 to 0.25.3 (#299)
- Bump github.com/stretchr/testify from 1.8.0 to 1.8.1 (#298)
- Bump github.com/bmatcuk/doublestar/v4 from 4.2.0 to 4.3.0 (#297)

## 2.5.4 2022-09-27

### Maintenance

- Remove puckpuck from dependabot reviewers (#283) | [@puckpuck](https://github.com/puckpuck)
- Bump github.com/bmatcuk/doublestar/v4 from 4.0.2 to 4.2.0 (#291)
- Bump dependencies (#286)
- Bump github.com/mitchellh/mapstructure from 1.4.3 to 1.5.0 (#254)

## 2.5.3 2022-07-25

- fixed openSSL CVE | [@pkanal](https://github.com/pkanal)

## 2.5.2 2022-06-09

### Fixes

- Only return cpu.utilization if a limit was provided (#262) | [@MikeGoldsmith](https://github.com/MikeGoldsmith)

### Maintenance

- Deflake flaky tests (#264) | [@kentquirk](https://github.com/kentquirk)
- Create helm-chart issue on release (#255) | [@MikeGoldsmith](https://github.com/MikeGoldsmith)
- Update codeowners to only be telemetry team (#256) | [@MikeGoldsmith](https://github.com/MikeGoldsmith)
- Bump github.com/honeycombio/honeytail from 1.6.1 to 1.6.2 (#253)

## 2.5.1 2022-04-08

### Maintenance

- Update to Go 1.18; fixes openSSL CVE (#237) | [@MikeGoldsmith](https://github.com/MikeGoldsmith)

## 2.5.0 2022-04-08

### Added

- Add exclude capability for certain logging paths (#241) | [@kentquirk](https://github.com/kentquirk)

### Maintenance

- Update memory.utilization memory counter to use memory.workingset (#242) | [@MikeGoldsmith](https://github.com/MikeGoldsmith)
- Bump k8s.io/kubelet from 0.23.3 to 0.23.5 (#244)
- Bump github.com/stretchr/testify from 1.7.0 to 1.7.1 (#245)
- Bump k8s.io/api from 0.23.3 to 0.23.5 (#246)

## 2.4.0 2022-03-02

### Added

- Implements retry logic for sending events (#232) | [@puckpuck](https://github.com/puckpuck)
- Unify logging, fix log verbosity (#229) | [@puckpuck](https://github.com/puckpuck)

### Fixed

- Return early on empty MemoryStats (#230) | [@JamieDanielson](https://github.com/JamieDanielson)
- Fix broken test from logrus change (#231) | [@JamieDanielson](https://github.com/JamieDanielson)

### Maintenance

- Bump k8s.io/kubelet from 0.23.1 to 0.23.3 (#224)
- Bump github.com/honeycombio/honeytail from 1.6.0 to 1.6.1 (#221)
- Bump k8s.io/client-go from 0.23.1 to 0.23.3 (#220)

## 2.3.2 2022-02-09

### Maintenance

- Bump github.com/honeycombio/libhoney-go from 1.15.7 to 1.15.8 (#223) | [dependabot](https://github.com/dependabot)

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
