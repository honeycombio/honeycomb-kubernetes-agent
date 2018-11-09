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

### Smoketesting multiple K8S versions with Minikube

Testing the agent against each version of Kubernetes is tedious, but hopefully this can streamline the process for you.

**Requirements**

- minikube
- virtualbox (assumes you have a VM driver)

#### Steps

**1. Launch the version of the k8s cluster you want**

```bash
$ cd smoketest
$ ./test-k8s.sh 1.10.1
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 51.3M  100 51.3M    0     0  3319k      0  0:00:15  0:00:15 --:--:-- 7245k
/Users/tredman/go/src/github.com/honeycombio/honeycomb-kubernetes-agent/smoketest
Deleting local Kubernetes cluster...
Machine deleted.
Starting local Kubernetes v1.10.1 cluster...
Starting VM...
Getting VM IP address...
Moving files into cluster...
Downloading kubeadm v1.10.1
Downloading kubelet v1.10.1
Finished Downloading kubelet v1.10.1
Finished Downloading kubeadm v1.10.1
Setting up certs...
Connecting to cluster...
Setting up kubeconfig...
Starting cluster components...
Kubectl is now configured to use the cluster.
Loading cached images from config file.
```

This will use minikube to launch the specified version in your virtualbox environment, and download the right kubectl for that version.

**2. Put the kubectl in your path temporarily**

```bash
$ source ./activate.sh
```

**3. Verify the k8s/kubectl versions are correct**

```bash
$ kubectl version
Client Version: version.Info{Major:"1", Minor:"10", GitVersion:"v1.10.1", GitCommit:"d4ab47518836c750f9949b9e0d387f20fb92260b", GitTreeState:"clean", BuildDate:"2018-04-12T14:26:04Z", GoVersion:"go1.9.3", Compiler:"gc", Platform:"darwin/amd64"}
Server Version: version.Info{Major:"1", Minor:"10", GitVersion:"v1.10.1", GitCommit:"d4ab47518836c750f9949b9e0d387f20fb92260b", GitTreeState:"clean", BuildDate:"2018-04-12T14:14:26Z", GoVersion:"go1.9.3", Compiler:"gc", Platform:"linux/amd64"}
```

**4. Install the k8s agent**

```bash
$ kubectl apply -f quickstart.yaml
serviceaccount "honeycomb-serviceaccount" created
clusterrolebinding.rbac.authorization.k8s.io "honeycomb-serviceaccount" created
clusterrole.rbac.authorization.k8s.io "honeycomb-serviceaccount" created
configmap "honeycomb-agent-config" created
daemonset.extensions "honeycomb-agent-v1.1" created
secret "honeycomb-writekey" created
```

**5. Optionally install the "guestbook-frontend" sample app**

This will give you a pod that the Honeycomb agent should pick up.

```bash
$ kubectl apply -f frontend-deployment.yaml
deployment.apps "frontend" created
$ kubectl get pods
NAME                         READY     STATUS              RESTARTS   AGE
frontend-5c548f4769-p6j8k    0/1       ContainerCreating   0          48s
honeycomb-agent-v1.1-8vdb9   0/1       ContainerCreating   0          3s
```

**6. Push your version of the agent to k8s**

```bash
# from the repo root
$ make container
building: bin/amd64/honeycomb-kubernetes-agent
Sending build context to Docker daemon  196.4MB
Step 1/5 : FROM golang:1.8-alpine
 ---> 4cb86d3661bf
Step 2/5 : MAINTAINER Team Honeycomb <bees@honeycomb.io>
 ---> Using cache
 ---> c2d6551e3573
Step 3/5 : ADD bin/amd64/honeycomb-kubernetes-agent /honeycomb-kubernetes-agent
 ---> 6ff387285b08
Step 4/5 : USER root:root
 ---> Running in b8e0ef6f3468
Removing intermediate container b8e0ef6f3468
 ---> 1d0609ade8fd
Step 5/5 : ENTRYPOINT ["/honeycomb-kubernetes-agent"]
 ---> Running in 8606e25648b0
Removing intermediate container 8606e25648b0
 ---> b6d13547c350
Successfully built b6d13547c350
Successfully tagged honeycombio/honeycomb-kubernetes-agent:5828ce8-dirty
container: honeycombio/honeycomb-kubernetes-agent:5828ce8-dirty

# now deploy it to your cluster
$ docker save honeycombio/honeycomb-kubernetes-agent:5828ce8-dirty | pv | (eval $(minikube docker-env) && docker load)
```

Update the quickstart.yaml with this development tag and redeploy:

```yaml
image: honeycombio/honeycomb-kubernetes-agent:CHANGEMETODEVTAG
```

```bash
$ kubectl apply -f quickstart.yaml
```

**7. Check the logs, poke around, verify your code, etc**

```bash
$ kubectl logs -l k8s-app=honeycomb-agent
time="2018-11-09T22:01:48Z" level=debug msg="Starting informer" fieldSelector="spec.nodeName=minikube" labelSelector="k8s-app=kube-controller-manager,k8s-app!=honeycomb-agent" namespace=kube-system
time="2018-11-09T22:01:48Z" level=debug msg="Starting informer" fieldSelector="spec.nodeName=minikube" labelSelector="app=guestbook,k8s-app!=honeycomb-agent" namespace=default
...
```
