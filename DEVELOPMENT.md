## Development Mechanics

You'll need Go and Docker installed (obviously). For local testing, you can use
[Minikube](https://kubernetes.io/docs/getting-started-guides/minikube/). See
notes below on tweaks needed for Minikube.

If you need a fuller-sized throwaway cluster, I recommend the [Heptio AWS
quickstart](https://s3.amazonaws.com/quickstart-reference/heptio/latest/doc/heptio-kubernetes-on-the-aws-cloud.pdf).
Just make sure to use a testing AWS account, or at least make sure that you don't
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

### Working with Docker Kubernetes 

Docker recently released a Kubernetes integration in Docker Edge. For information on how to install and setup [go
here](https://docs.docker.com/docker-for-mac/kubernetes/). 

Luckily, deploying our spec file is much easier than setting up minikube! To build the local image:

```
 $ make container
building: bin/amd64/honeycomb-kubernetes-agent
Sending build context to Docker daemon  156.1MB
...
container: honeycombio/honeycomb-kubernetes-agent:1fb43a1-dirty
```

Verify that it published to your local docker images with:
```
$ docker images
REPOSITORY                                               TAG                 IMAGE ID            CREATED             SIZE
honeycombio/honeycomb-kubernetes-agent                   1fb43a1-dirty       a2b0e38e3a85        5 minutes ago       302MB
```

#### If it failed to publish to your local docker repo, try building the container manually with:

```
docker build -t honeycomb-kubernetes-agent:{$TAG} -f .dockerfile-amd64 .
```

Now it's time to update the example spec file located in `examples/quickstart.yaml`, we need to add which image 
and it's _paramount_ to set `imagePullPolicy: IfNotPresent`. Setting it to `Always` causes kubectl to not pull from your
local Docker repository.

```
      - env:
        - name: HONEYCOMB_WRITEKEY
          valueFrom:
            secretKeyRef:
              key: key
              name: honeycomb-writekey
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        image: honeycombio/honeycomb-kubernetes-agent:{$TAG}
        imagePullPolicy: IfNotPresent 
```


From here, we can create a deploy on your locally running Kubernetes cluster using kubectl:
```
 $ kubectl apply -f examples/quickstart.yaml
serviceaccount "honeycomb-serviceaccount" created
clusterrolebinding "honeycomb-serviceaccount" created
clusterrole "honeycomb-serviceaccount" created
configmap "honeycomb-agent-config" created
daemonset "honeycomb-agent-v1.1" created
```

Confirm that the pods are up with:
```
 $ kubectl get pods --namespace=kube-system
NAME                                         READY     STATUS    RESTARTS   AGE
etcd-docker-for-desktop                      1/1       Running   1          5d
honeycomb-agent-v1.1-pzk68                   1/1       Running   0          11m
kube-apiserver-docker-for-desktop            1/1       Running   1          5d
kube-controller-manager-docker-for-desktop   1/1       Running   1          6d
kube-dns-6f4fd4bdf-jggsq                     3/3       Running   0          5d
kube-proxy-8zpt4                             1/1       Running   0          5d
kube-scheduler-docker-for-desktop            1/1       Running   1          5d
```

To view the events and describe the pod you can run, this has a lot of useful information:
```
$ kubectl describe pods --namespace=kube-system --selector=k8s-app=honeycomb-agent
```

To view logs you can run:
```
kubectl logs --namespace=kube-system --selector=k8s-app=honeycomb-agent
```

##### For iterative local changes

If you're continually developing on the binary, to have kubectl reapply your image to the running pod, you need
to delete the deploy. 

```
 $ kubectl delete -f examples/quickstart.yaml
serviceaccount "honeycomb-serviceaccount" deleted
clusterrolebinding "honeycomb-serviceaccount" deleted
clusterrole "honeycomb-serviceaccount" deleted
configmap "honeycomb-agent-config" deleted
daemonset "honeycomb-agent-v1.1" deleted
```

Once you've done that, rebuild your container image and reapply the spec! 







