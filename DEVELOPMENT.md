## Development Mechanics

You'll need Go and Docker installed (obviously). For local testing, you can use
[Minikube](https://kubernetes.io/docs/getting-started-guides/minikube/). See
notes below on tweaks needed for Minikube.

If you need a fuller-sized throwaway cluster, I recommend the [Heptio AWS
quickstart](https://s3.amazonaws.com/quickstart-reference/heptio/latest/doc/heptio-kubernetes-on-the-aws-cloud.pdf)
Just mmake sure to use a testing AWS account, or at least make sure that you don't
interfere with an existing VPC setup.

To build the agent image, run `make container`; to run unit tests and `go vet`,
run `make test`.

CI runs an additional end-to-end smoke test that sets up a Minikube cluster and
sends events through it. You can find that in the `e2e-tests/` directory.

### Working with Minikube

To test inside Minikube with a locally-built image:
1. Modify the DaemonSet spec to mount `/mnt/sda1/var/lib/docker/containers` as a `volumeMount`. This is because `/var/lib/docker/containers` is symlinked to `/mnt/sda1/var/lib/docker/containers` in the Minikube VM.

2. Make sure that you specify `imagePullPolicy: IfNotPresent` or `imagePullPolicy: Never` in the container spec.

3. To make the local container image inside Minikube, run `make container`, then `docker save honeycombio/honeycomb-kubernetes-agent:$TAG | minikube ssh docker load`.

(Alternative strategies for step 3 may be possible; see the [minikube docs](https://github.com/kubernetes/minikube/blob/master/docs/reusing_the_docker_daemon.md) for more details on building local images, and [this blog post](https://blog.hasura.io/sharing-a-local-registry-for-minikube-37c7240d0615) on sharing a local container registry.)
