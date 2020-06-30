# Nuxeo Operator

This project is a very early - **0.4.0 at present** - OpenShift/Kubernetes Operator written in Go to manage the state of a Nuxeo cluster. Nuxeo is an open source content management system. (See https://www.nuxeo.com/). The Operator scaffolding was initially generated using the Operator SDK (https://docs.openshift.com/container-platform/4.4/operators/operator_sdk/osdk-getting-started.html/).

Presently, I'm doing this development on a Ubuntu 18.04 desktop with OpenShift Code Ready Containers (https://github.com/code-ready/crc) and MicroK8s (https://microk8s.io).

### Planned Features

Below is the sequence of capabilities that are planned for this Operator. These are all preliminary and will be tuned as I get further into the project.

#### Version 0.1.0 *(complete)*

This version is really just a POC of the Operator with basic functionality. The goal is to be able to bring up - and reconcile the state of - a basic Nuxeo cluster that optionally supports TLS via an Nginx reverse proxy. The goal is to do this manually, and also via an OLM subscription.

This version creates/reconciles a Deployment, a Service, a Service Account, and an OpenShift Route. It also includes the ability to install Nuxeo Marketplace packages (https://connect.nuxeo.com/nuxeo/site/marketplace) assuming Internet access to the Marketplace by the Operator.

This version of the Operator is pre-Level I capability. (https://sdk.operatorframework.io/docs/operator-capabilities/)

| Feature                                                      | Status      |
| ------------------------------------------------------------ | ----------- |
| Define an initial Nuxeo Custom Resource Definition (CRD) that supports the feature set for this increment of functionality | complete |
| Generate and reconcile a Deployment from a Nuxeo Custom Resource (CR) to represent the desired state of a Nuxeo cluster | complete |
| Run Nuxeo only in development mode with only embedded services | complete |
| Support *Pod Templates* in the Nuxeo CR for fine-grained configuration of Pods. Use a default Template if no Pod Template is provided | complete |
| Support an optional Nginx TLS reverse proxy sidecar container in each Nuxeo Pod to support TLS. Test using 443 outside the cluster, Nginx listening on 8443, and forwarding to Nuxeo on 8080. Model the sidecar configuration from the Nuxeo APB catalog (https://github.com/nuxeo/nuxeo-apb-catalog/blob/master/nuxeo-apb/files/nginx.conf) | complete |
| Create and reconcile a Route resource for access outside of the OpenShift cluster. Support only TLS passthrough at this time | complete |
| Create and reconcile a Service resource for the Route to use, and for potential use within the cluster. The service will communicate with Nuxeo on 8080, or Nginx on 8443 | complete |
| Create all resources that originate from a Nuxeo CR with `ownerReferences` that reference the Nuxeo CR - so that deletion of the Nuxeo CR will result in recursive removal of all generated resources for proper clean up | complete |
| Support custom Nuxeo images, with a default of `nuxeo:latest` if no custom image is provided in the Nuxeo CR | complete |
| Support running the Operator binary externally to the cluster to verify the basic Operator functionality | complete |
| Implement a minimal Status field of the Nuxeo CR consisting of the number of active pods  | complete |
| Perform basic testing whereby cluster resources are modified and expected resource reconciliation is performed by the Nuxeo Operator | complete |
| Add the basic elements (CSV, RBACs, bundling, etc.) to support packaging the Operator as a community Operator | complete |
| Deploy and test the Operator from an internal Operator registry in the cluster via OLM subscription | complete |
| Automate the build with *Make*                               | complete |



#### Version 0.2.0 *(complete)*

Version 0.2.0 lays the groundwork for productionalization by implementing initial basic unit testing, and end-to-end (e2e) testing, and extending the build automation to include these. The goal of the testing at this stage is to frame out the approach, select the test tooling, and create an initial set of reasonable tests. In order to be fully production-ready, more tests will clearly be needed, and will be added as work progresses.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Incorporate unit testing into the operator build using https://github.com/stretchr/testify | complete |
| Incorporate end-to-end testing, cloned from the operator-sdk memcached test. | complete |
| Integrate unit and e2e tests into the Make file | complete |



#### Version 0.3.0 *(complete)*

Version 0.3.0 adds Kubernetes support - testing with MicroK8s only at this time.

| Feature                                                      | Status   |
| ------------------------------------------------------------ | -------- |
| Implement the ability to detect whether the Operator is running in a Kubernetes cluster vs. an OpenShift cluster | complete |
| Create an *Ingress* resource for access outside of the Kubernetes cluster and test with HTTP as well as TLS passthrough | complete |
| Support a suitable Dockerfile for the Operator (Operator SDK generates a Dockerfile based on Red Hat *ubi*) | complete |
| Document Kubernetes testing using MicroK8s (https://microk8s.io/) and kubectl | complete |
| Test with the MicroK8s built-in registry (https://microk8s.io/docs/registry-built-in) for equivalence with the OpenShift internal image registry | complete |
| Update the Make file to support both OpenShift/CRC *and* MicroK8s | complete |



#### Version 0.4.0 *(complete)*

Version 0.4.0 incorporates additional basic features into the Operator.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Support readiness and liveness probes for the Nuxeo pods | complete |
| Support storage configuration for Nuxeo binaries, the transient store, etc. | complete  |
| Support explicit definition of nuxeo.conf properties in the Nuxeo CR | complete |
| Support additional fields in the Nuxeo CR as defined in the APB: Java opts, templates, packages, nuxeo URL, nuxeo name | complete |



#### Version 0.5.0 *(in progress)*

Version 0.5.0 incorporates some more advanced features into the Operator and extends the unit testing.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Support a Secret with payload for TLS termination in the Route/Ingress. Previously, TLS passthrough was the only tested functionality |        |
| Support the ability to terminate TLS directly in Nuxeo, rather than requiring a sidecar. | complete  |
| Support a secret for JVM-wide PKI configuration in the Nuxeo Pod - in order to support cases where Nuxeo is running in a PKI-enabled enterprise and is interacting with internal PKI-enabled Corporate micro-services that use an internal corporate CA. This would require a CA bundle+cert and passwords | complete |
| Support installing marketplace packages in disconnected mode if no Internet connection is available in-cluster | complete |
| Ability to configure *Interactive* nodes and *Worker* nodes differently. The objective is to support compute-intensive back-end processing on a set of nodes having a greater resource share in the cluster then the interactive nodes that serve the Nuxeo GUI |        |
| Support clustering - use Pod UID as `nuxeo.cluster.nodeid` via the downward API | complete |
| Support defining resource request/limit in the Nuxeo CR |          |
| Support Nuxeo CLID |          |



#### Version 0.6.0

Version 0.6.0 supports the *Service Binding Operator* to facilitate integration of a Nuxeo Cluster with backing services such as PostgreSQL, Kafka, and ElasticSearch.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Integrate with the Service Binding Operator (https://github.com/redhat-developer/service-binding-operator) to bind Nuxeo to various backing services present in the cluster |        |
| Test with Strimzi (https://strimzi.io/) for Nuxeo Stream support |        |
| Test with Crunchy PostgreSQL (https://www.crunchydata.com/products/crunchy-postgresql-for-kubernetes/) for database support |        |
| Test with Elastic Cloud on Kubernetes (https://github.com/elastic/cloud-on-k8s) for ElasticSearch support |        |
| Support password changes in backing services that trigger rolling updates of the Nuxeo cluster |        |
| Support certificate expiration and replacement for the JVM, and also for backing services that trigger rolling updates of the Nuxeo cluster. An example would be where the Kafka backing service is accessed via TLS, and the Kafka CA and cert expire and are renewed |        |



#### Version 0.7.0

Version 0.7.0 makes the Operator available as a Community Operator.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Build out unit tests for close to 100% coverage. Extend unit tests to cover more scenarios associated with various mutations of the Nuxeo CR - adding then removing then adding, etc. to ensure the reconciliation logic is robust |  |
| Address all "todo-me" in the code |  |
| Gain access to a full production-grade OpenShift cluster, and a full production-grade Kubernetes cluster to ensure compatibility with those production environments |        |
| Build out the Status field in the Nuxeo CR to be comparable with other resources available on OperatorHub | |
| Develop and test the elements needed to qualify the Operator for evaluation as a community Operator. Submit the operator for evaluation. Iterate |        |
| Provide `kustomize` examples to illustrate bringing up an exemplar Nuxeo Cluster using kustomize) https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/) |        |
| Review and augment unit and e2e tests                        |        |
| Support multi-architecture build. Incorporate lint, gofmt, etc. into the build process |        |
| GitHub CI / badges |        |
| Review the license |        |
| Find someone else to work on this with... | |
| Make the Operator available as a community Operator (https://github.com/operator-framework/community-operators) |        |



#### Other...

These have not been prioritized yet.

| Feature                         | Status |
| ------------------------------- | ------ |
| Phase V Operator Maturity Model |        |
| OperatorHub availability |        |
| Deploy a cluster as a Stateful Set or Deployment |        |
| JetStack Cert Manager integration |        |
| Horizontal Pod Auto-scaling |        |
| cert-utils support? (https://github.com/redhat-cop/cert-utils-operator) | |
| Other?... |        |



------

## Testing the Operator

### Testing with manual deployment

These are the steps to test the Operator via manual deployment. In the next section, instructions are provided for testing via OLM Subscription.

Testing has been performed using *Code Ready Containers* (CRC) for OpenShift, and Canonical *MicroK8s* (MK8s) for Kubernetes.

***Before following these test steps, please see the `docs` directory for important configuration information for CRC and MK8s.***

A later release will document instructions for running in a full OpenShift or Kubernetes cluster. But for now, if you're testing in a full cluster you will have to fill in the blanks yourself. In these instructions, when there is a difference in the steps for CRC vs MicroK8s, both instructions are provided.

##### A general note about these test steps

Before progressing, hopefully you've read the CRC or MK8s README as discussed above. If you're running on MK8s, it was recommended that you create an alias `mkubectl` for the MK8s `kubectl`. In the documentation that follows, the notation `(m)kubectl` refers to the `mkubectl` alias if running MK8s, and `kubectl` if running CRC. Of cource the CRC-provided `oc` command works as well.

##### Steps

Since the project is still in an early development stage, built-in image registries are used for storing images. This simplifies testing because these dev images will always be available in-cluster. for OpenShift/CRC, create an `images` namespace to hold various images/image streams: 

```shell
# CRC only
$ kubectl create namespace images
```

Create a Nuxeo image:

```shell
$ docker pull nuxeo:10.10
# CRC
$ REGISTRY=$(kubectl get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
$ docker login -u kubeadmin -p $(oc whoami -t) $REGISTRY
# MK8s
$ REGISTRY=localhost:32000
# common
$ docker tag nuxeo:10.10 $REGISTRY/images/nuxeo:10.10
$ docker push $REGISTRY/images/nuxeo:10.10
```

Deploy the Nuxeo CRD into the cluster:

```shell
# MK8s - on occasion MK8s seems to hang on re-apply so if re-applying, delete first
$ mkubectl delete crd/nuxeos.nuxeo.com
$ (m)kubectl apply -f deploy/crds/nuxeo.com_nuxeos_crd.yaml
```

Create a `nuxeo` namespace to test the Operator in, and for CRC grant all service accounts in the `nuxeo` namespace the ability to pull images from the `images` namespace:
```shell
$ (m)kubectl create namespace nuxeo
# CRC
$ oc policy add-role-to-group system:image-puller system:serviceaccounts:nuxeo\
  --namespace=images
```

Create a Nuxeo CR for a single-node Nuxeo instance with port 80 access external to the cluster. Note the different path for the MK8s resource due to the different image registry reference.
```shell
# CRC
$ kubectl apply -f deploy/examples/nuxeo-cr.yaml -n nuxeo
# MK8s
$ mkubectl apply -f hack/microk8s/nuxeo-cr-mk8s.yaml -n nuxeo
```

Build the Operator binary. This step assumes the current working directory is the directory into which you git cloned this project. Note - this project is Go 1.14. Use the provided Make target:

```shell
$ go version
go version go1.14.2 linux/amd64
$ make operator-build
```

Run the Operator outside of the cluster from the command line. To run the operator this way, you provide the  watch namespace as an environment variable, and a command-line option specifying the path of a kube config with credentials for the cluster. For MK8s use the kube config in its snap location:

```shell
# CRC
$ WATCH_NAMESPACE=nuxeo build/_output/bin/nuxeo-operator --kubeconfig=$HOME/.kube/config
# MK8s:
WATCH_NAMESPACE=nuxeo build/_output/bin/nuxeo-operator --kubeconfig=/var/snap/microk8s/current/credentials/kubelet.config
```

You should get output like the following displayed to the console:

```shell
{"level":"info","ts":1591372655.1064653,"logger":"cmd","msg":"Operator Version: 0.0.1"}
{"level":"info","ts":1591372655.1065361,"logger":"cmd","msg":"Go Version: go1.14.2"}
{"level":"info","ts":1591372655.1065524,"logger":"cmd","msg":"Go OS/Arch: linux/amd64"}
{"level":"info","ts":1591372655.1065633,"logger":"cmd","msg":"Version of operator-sdk: v0.17.1"}
...
(remainder deleted for brevity)
```

The operator will run until you press CTRL-C to stop it, so leave it running in this console window for the remainder of this test and open a new console window to work with. Verify that the Nuxeo Operator created the expected resources:

```shell
# CRC
$ kubectl get service,sa,deployment,replicaset,route,pod -n nuxeo
# MK8s:
$ mkubectl get service,sa,deployment,replicaset,ingress,pod -n nuxeo
# CRC output is shown
NAME                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   AGE
service/my-nuxeo-cluster-service   ClusterIP   172.30.40.129   <none>        80/TCP    32s

NAME                      SECRETS   AGE
...
serviceaccount/nuxeo      2         32s

NAME                                     READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/my-nuxeo-cluster   1/1     1            1           32s

NAME                                                DESIRED   CURRENT   READY   AGE
replicaset.extensions/my-nuxeo-cluster-6c4b7466dc   1         1         1       32s

NAME                                              HOST/PORT                     ...
route.route.openshift.io/my-nuxeo-cluster-route   nuxeo-server.apps-crc.testing ...

NAME                                    READY   STATUS    RESTARTS   AGE
pod/my-nuxeo-cluster-6c4b7466dc-zc4wh   1/1     Running   0          32s

```

Test access to Nuxeo. Of key importance here is that the host name in the route - `nuxeo-server.apps-crc.testing` - is addressable from your local host. The CRC and MicroK8s setup documents in the docs folder describe this.

```shell
# CRC
$ ping nuxeo-server.apps-crc.testing
PING nuxeo-server.apps-crc.testing (192.168.130.11) 56(84) bytes of data.
64 bytes from api.crc.testing (192.168.130.11): icmp_seq=1 ttl=64 time=0.333 ms
...
# MK8s
$ ping nuxeo-server.apps-crc.testing
PING nuxeo-server.apps-crc.testing (127.0.0.1) 56(84) bytes of data.
64 bytes from localhost (127.0.0.1): icmp_seq=1 ttl=64 time=0.075 ms
# common
$ curl http://nuxeo-server.apps-crc.testing
<META HTTP-EQUIV="refresh" CONTENT="0;URL=/nuxeo">
```

Assuming success so far, then it should be possible to access the Nuxeo Pod from your browser using the same HTTP URL shown in the curl example above. If everything worked you should see the Nuxeo main page:

![](resources/images/nuxeo-ui.jpg)

To review what happened so far:

1. You deployed a Nuxeo CR that specified the `nuxeo-web-ui` Marketplace package
2. You ran the operator from your desktop
3. The Operator saw the Nuxeo CR you generated, and created a Route/Ingress, a Service, a Service Account, and a Deployment. The Operator passed the `NUXEO_PACKAGES` environment variable from the Nuxeo CR into the Deployment
4. The cluster reconciled the Deployment into a ReplicaSet and a Pod. The Pod also got the `NUXEO_PACKAGES` environment variable. The Pod runs under the generated `nuxeo` Service Account
5. The Pod started the Nuxeo Container
6. The Nuxeo Container saw the environment variable `NUXEO_PACKAGES=nuxeo-web-ui` and so it installed that package into the Nuxeo Container
7. You should now be able to log into this unlicensed development version of Nuxeo as `Administrator/Administrator` (make sure cookies are enabled)

##### **Next, test TLS access using an Nginx sidecar**

Create an *Ningx* image in the `images` namespace. This step uses the `$REGISTRY` environment variable defined above:

```shell
$ echo $REGISTRY
# CRC (yours might be different):
default-route-openshift-image-registry.apps-crc.testing
# MK8s:
localhost:32000
# common
$ docker pull nginx:latest
$ docker tag nginx:latest $REGISTRY/images/nginx:latest
$ docker push $REGISTRY/images/nginx:latest
latest: digest: sha256:... size: 1362
# CRC
$ kubectl get imagestreams -n images
NAME    IMAGE REPOSITORY                                                      TAGS
nginx   default-route-openshift-image-registry.apps-crc.testing/nuxeo/nginx   latest
nuxeo   default-route-openshift-image-registry.apps-crc.testing/nuxeo/nuxeo   10.10
# MK8s
$ curl http://localhost:32000/v2/_catalog
{"repositories":["images/nginx","images/nuxeo"]}
$ curl http://localhost:32000/v2/images/nginx/tags/list
{"name":"images/nginx","tags":["latest"]}
$ curl http://localhost:32000/v2/images/nginx/manifests/latest
{
   "schemaVersion": 1,
   "name": "images/nginx",
   "tag": "latest",
   "architecture": "amd64",
   "fsLayers": [
   ...
# etc...
```


Generate a self-signed TLS certificate and key, and a `dhparams` file to use in terminating the TLS connection in the Nginx sidecar. You can use the script `hack/make-tls-secret` to create a secret YAML in the `tmp` directory, or you can do it by hand:
```shell
$ mkdir tmp
$ cd tmp
$ openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 \
 -keyout nginx-selfsigned.key \
 -out nginx-selfsigned.crt \
 -subj "/C=US/ST=Maryland/L=Somewhere/O=IT/CN=nuxeo-server.apps-crc.testing"
$ openssl dhparam -out nginx-dhparam-2048.pem 2048
$ ls
nginx-dhparam-2048.pem  nginx-selfsigned.crt  nginx-selfsigned.key
```

Create a TLS Secret YAML manifest from the items above. This Secret will be used by the Nuxeo Operator further on down. This assumes the `tmp` directory is still the active directory where the three files above were generated. Just paste this into your command terminal:
```shell
rm ./tls-secret.yaml
cat <<EOF > tls-secret.yaml
apiVersion: v1
kind: Secret
type: kubernetes.io/tls
metadata:
  name: tls-secret
stringData:
  tls.key: |
EOF
sed -e 's/^/    /' nginx-selfsigned.key >> tls-secret.yaml 
echo '  tls.crt: |' >> tls-secret.yaml 
sed -e 's/^/    /' nginx-selfsigned.crt >> tls-secret.yaml 
echo '  dhparam: |' >> tls-secret.yaml 
sed -e 's/^/    /' nginx-dhparam-2048.pem >> tls-secret.yaml
```

This should produce a secret file named `tls-secret.yaml` like:

```shell
$ cat tls-secret.yaml 
apiVersion: v1
kind: Secret
type: kubernetes.io/tls
metadata:
  name: tls-secret
stringData:
  tls.key: |
    -----BEGIN PRIVATE KEY-----
    MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCyPF7vM37+g56c
    ...
   1g3UPjfIXOXlAU7JPPhe46Ae
    -----END PRIVATE KEY-----
  tls.crt: |
    -----BEGIN CERTIFICATE-----
    MIIDszCCApugAwIBAgIUTdDP8BjurjtNpP599D+39J3fd4MwDQYJKoZIhvcNAQEL
    ...
    aImWX5zkHru8kJZwrrJ636Z9FFbCu0uZej8nPgJYomRCrYYD2sLP
    -----END CERTIFICATE-----
  dhparam: |
    -----BEGIN DH PARAMETERS-----
    MIIBCAKCAQEAgCfJKgqUs1SZzKDwSytVBIkB8zCbBskZN9OBkXjR/+ZBlIcN5FKm
    ...
    KAVAbSJxz/Wm3sOKpjXr+esAMHqPsRXXYwIBAg==
    -----END DH PARAMETERS-----
```

Now, generate the Secret you created above, as well as a ConfigMap provided with the Operator. Both of these resources are used by the Operator to configure the Nuxeo pod Nginx sidecar. This assumes you are still in the `tmp` directory:

```shell
$ (m)kubectl apply -f tls-secret.yaml,../deploy/examples/tls-cfgmap.yaml -n nuxeo
$ (m)kubectl get secret,cm -n nuxeo
NAME                              TYPE                                  DATA   AGE
...
secret/tls-secret                 kubernetes.io/tls                     3      84s

NAME                   DATA   AGE
configmap/tls-cfgmap   2      44s
```

Now, create a different Nuxeo CR - this one configures Nuxeo for port 443 access outside the cluster. This CR has a different spec that indicates Nginx is to be used as a reverse proxy:

```shell
$ cd ..
# CRC
$ kubectl apply -f deploy/examples/nuxeo-cr-tls.yaml -n nuxeo
# MK8s
$ mkubectl apply -f hack/microk8s/nuxeo-cr-tls-mk8s.yaml -n nuxeo
```

Observe the Operator modifying the Service, and the Deployment. Observe the Deployment regenerating the Pod. Wait for the Pod to stabilize. Note that the Nuxeo Pod now contains two Containers (as indicated by a Ready state of 2/2). The first Container is the Nginx sidecar, and the second Container is Nuxeo:
```shell
$ (m)kubectl get pod -n nuxeo
NAME                                READY   STATUS    RESTARTS   AGE
my-nuxeo-cluster-84b7bc487b-h2rdz   2/2     Running   0          40s
```

You should now be able to access Nuxeo via HTTPS by pasting the Nuxeo URL into your browser:
```shell
https://nuxeo-server.apps-crc.testing/nuxeo
```

Because Nginx is terminating TLS with a self-signed certificate, your browser will display a certificate authenticity warning. Just go ahead and add an exception. That should take you to exactly the same page as you saw before. Check the Nginx container logs to verify that it is acting as the reverse proxy:

```shell
$ (m)kubectl logs my-nuxeo-cluster-84b7bc487b-h2rdz -c nginx -f -n nuxeo
/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
...
{"time_local":"05/Jun/2020:17:20:35 +0000","remote_addr":"10.128.0.1","remote_user":"","request":"GET /nuxeo HTTP/1.1","status": "302","body_bytes_sent":"5","request_time":"0.045","http_referrer":"","http_user_agent":"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:76.0) Gecko/20100101 Firefox/76.0"}
...
```

Go back to the console that was running the Nuxeo Operator from the command line and press CTRL-C to stop the Operator. You should be returned to the command prompt.

Review the resources created by the Nuxeo Operator. Then, delete the Nuxeo CR and observe that all of those resources are cleaned up automatically by the cluster via cascading delete, because of the `ownerReferences` on the resources:

```shell
# CRC
$ kubectl get nuxeo,service,sa,deployment,replicaset,route,pod -n nuxeo
# MK8s:
$ mkubectl get nuxeo,service,sa,deployment,replicaset,ingress,pod -n nuxeo
# CRC output is shown
NAME                       AGE
nuxeo.nuxeo.com/my-nuxeo   106m

NAME                                     READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/my-nuxeo-cluster   1/1     1            1           103m

NAME                                                DESIRED   CURRENT   READY   AGE
replicaset.extensions/my-nuxeo-cluster-84b7bc487b   1         1         1       40m

NAME                                    READY   STATUS    RESTARTS   AGE
pod/my-nuxeo-cluster-84b7bc487b-h2rdz   2/2     Running   0          21m

NAME                                              HOST/PORT                     ...
route.route.openshift.io/my-nuxeo-cluster-route   nuxeo-server.apps-crc.testing ...

NAME                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S) ...
service/my-nuxeo-cluster-service   ClusterIP   172.30.80.128   <none>        443/TCP ...

NAME                      SECRETS   AGE
...
serviceaccount/nuxeo      2         103m

$ (m)kubectl delete nuxeo my-nuxeo
```

Allow a moment or two for the cluster to remove the resources. Then:

```shell
$ (m)kubectl get nuxeo,deployment,replicaset,pod... -n nuxeo
No resources found in nuxeo namespace.
$ (m)kubectl get sa/nuxeo -n nuxeo
Error from server (NotFound): serviceaccounts "nuxeo" not found
```



------

### Testing the Operator via OLM

These are the steps to test the Nuxeo Operator via OLM. If you didn't execute the prior section, then create an `images` namespace and put the Nuxeo 10.10 docker image into the namespace:

```shell
$ (m)kubectl create namespace images
$ docker pull nuxeo:10.10
# CRC
$ REGISTRY=$(kubectl get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
$ docker login -u kubeadmin -p $(oc whoami -t) $REGISTRY
# MK8s
$ REGISTRY=localhost:32000
# common
$ docker tag nuxeo:10.10 $REGISTRY/images/nuxeo:10.10
$ docker push $REGISTRY/images/nuxeo:10.10
```

Build and deploy the Nuxeo Operator OLM *Index* - which is what OLM uses to instantiate the operator via an OLM `subscription`. This section will use the Make file for these tasks. The Make targets that accomplish this are:

1. **operator-build** - build the operator Go binary from Go sources
2. **operator-image-build** - build the operator container image
3. **operator-image-push** - push the operator container image to the cluster
4. **bundle-build** - build the Operator bundle image
7. **index-add** - generate an OLM Index, which references the bundle image
8. **index-push** - push the OLM Index to the cluster

Because this is still an early development project, all images are pushed by the Make file into the cluster. The Make file uses the *operator-sdk* command to implement some of the OLM integration. The Make file requires operator-sdk **v0.18.0** so you need to make sure you're running that version. The Make file also uses the **opm** command. You need to build that if it is not already available:

```shell
$ pushd $HOME/go
$ git clone https://github.com/operator-framework/operator-registry.git
$ cd operator-registry
$ make
$ sudo cp bin/opm /usr/local/bin/opm
$ popd
```

Generate the Nuxeo Operator image. This will compile the Go binary, and push the resulting image into the `images` namespace. This step assumes your working directory is project root, e.g. `$HOME/go/nuxeo-operator`:

```shell
# CRC
$ make operator-build operator-image-build operator-image-push
# MK8s
$ TARGET_CLUSTER=MICROK8S make operator-build operator-image-build operator-image-push
```

If success , then you will see the Operator image in the `images` namespace, along with the Nuxeo image you pushed previously:

```shell
# CRC
$ kubectl get is -n images
NAME             IMAGE REPOSITORY    TAGS    UPDATED
nuxeo            ...nuxeo            10.10   6 minutes ago
nuxeo-operator   ...nuxeo-operator   0.x.0   35 seconds ago
# MK8s
$ curl http://localhost:32000/v2/_catalog
{"repositories":["images/nginx","images/nuxeo","images/nuxeo-operator"]}
```

Next, generate the OLM components to enable the Operator to be deployed via OLM. For OpenShift, the Make file generates images in the `custom-operators` namespace. (In MK8s, the images are managed by the registry add-on and the concept of image streams does not apply.) Create the namespace:

```shell
# CRC only
$ kubectl create namespace custom-operators
```

Now, run the Make file to create the OLM Index that serves the Nuxeo Operator. This step assumes your working directory is project root, and builds on the prior Make step:

```shell
# CRC
$ make bundle-build index-add index-push
# MK8s
$ TARGET_CLUSTER=MICROK8S make bundle-build index-add index-push
```

If successful, then:

```shell
# CRC
$ kubectl get is -n custom-operators
NAME                             IMAGE REPOSITORY                                TAGS  ...
nuxeo-operator-index             default-route-openshift-image-registry.apps...  0.x.0 ...
nuxeo-operator-manifest-bundle   default-route-openshift-image-registry.apps...  0.x.0 ...
# MK8s
$ curl http://localhost:32000/v2/_catalog
{"repositories":["custom-operators/nuxeo-operator-index","custom-operators/nuxeo-operator-manifest-bundle","images/nginx","images/nuxeo","images/nuxeo-operator"]}
```


You can now use the Operator's OLM integration to instantiate a Nuxeo cluster. The summary steps are:

1. Create a new namespace for the Nuxeo cluster
2. Create a `CatalogSource` in the namespace to serve the OLM registry components in the namespace
3. Subscribe the Nuxeo Operator via OLM
4. Wait for the Operator to be running
5. Create a Nuxeo CR
6. Wait for the various Nuxeo objects to be created by the Operator
7. Access Nuxeo from your browser
8. Tear down the Nuxeo cluster
9. Remove the namespace

These steps are now presented in detail:

Create a namespace, and for OpenShift grant the namespace authorization to pull images from the `images` namespace **and** the `custom-operators` namespace:

```shell
$ (m)kubectl create namespace nuxeo-test
# CRC only
$ oc policy add-role-to-group system:image-puller system:serviceaccounts:nuxeo-test\
  --namespace=images
# you may get a warning like this - ignore it
Warning: Group 'system:serviceaccounts:nuxeo-test' not found
clusterrole.rbac.authorization.k8s.io/system:image-puller added: "system:serviceaccounts:nuxeo-test"

$ oc policy add-role-to-group system:image-puller system:serviceaccounts:nuxeo-test\
  --namespace=custom-operators
```

Create a `CatalogSource` in the `nuxeo-test` namespace to serve the OLM registry components in the namespace. Note that the steps below use a SHA reference because if you build the operator index multiple times, the OLM caching behavior won't find newer images if the tag doesn't change. So using a SHA isn't mandatory the first time, but following this process always works. And - note that the catalog source pulls an image from the custom-operators namespace and so the image puller role created above is mandatory:

```shell
# CRC
$ REGISTRY=$(make print-IMAGE_REGISTRY_CLUST);\
  VERSION=$(make print-OPERATOR_VERSION);\
  echo $REGISTRY $VERSION
# e.g. image-registry.openshift-image-registry.svc.cluster.local:5000 0.4.0
$ SHA=$(kubectl get is nuxeo-operator-index -n custom-operators -o\
  jsonpath='{@.status.tags[0].items[0].image}')

# MK8s
$ REGISTRY=$(TARGET_CLUSTER=MICROK8S make print-IMAGE_REGISTRY_CLUST);\
  VERSION=$(TARGET_CLUSTER=MICROK8S make print-OPERATOR_VERSION);\
  echo $REGISTRY $VERSION
# e.g. localhost:32000 0.4.0
$ SHA=$(curl -vH "Accept: application/vnd.docker.distribution.manifest.v2+json"\
  http://$REGISTRY/v2/custom-operators/nuxeo-operator-index/manifests/$VERSION 2>&1\
  | grep Docker-Content-Digest | awk '{print $3}')

# common
$ SHA=${SHA%$'\r'}; echo $SHA
# e.g.
sha256:22621b4ea0745b10df4a1e2d985c6b00d74b8a0d16ac42dac4001fcdc3d36d6c

$ cat <<EOF | (m)kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: nuxeo-operator-index
  namespace: nuxeo-test
spec:
  displayName: Nuxeo Operator Index
  publisher: Test
  sourceType: grpc
  image: $REGISTRY/custom-operators/nuxeo-operator-index@$SHA
EOF
```

Confirm the `CatalogSource` is operating correctly. It may take a moment for the pod to reach the *Running* state:

```shell
$ (m)kubectl get catalogsource,pod -n nuxeo-test
NAME                   DISPLAY                TYPE   PUBLISHER   AGE
nuxeo-operator-index   Nuxeo Operator Index   grpc   Test        29s

NAME                         READY   STATUS    RESTARTS   AGE
nuxeo-operator-index-k4r8p   1/1     Running   0          29s

$ (m)kubectl logs nuxeo-operator-index-k4r8p -n nuxeo-test
time="2020-06-09T18:04:12Z" level=info msg="serving registry" database=/database/index.db port=50051

$ (m)kubectl get packagemanifest nuxeo-operator -n nuxeo-test
NAME             CATALOG                AGE
nuxeo-operator   Nuxeo Operator Index   71s
```

If the Nuxeo OLM Index Pod was correctly built and is running successfully, it will be hosting a gRPC server serving on port 50051. The logs will be a single line as shown above.

Next, create an Operator Group, and then subscribe to the Nuxeo Operator via OLM:

```shell
$ cat <<EOF | (m)kubectl apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: operator-group
  namespace: nuxeo-test
spec:
  targetNamespaces:
  - nuxeo-test
---
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: nuxeo-operator
  namespace: nuxeo-test
spec:
  channel: alpha
  name: nuxeo-operator
  source: nuxeo-operator
  sourceNamespace: nuxeo-test
EOF
operatorgroup.operators.coreos.com/operator-group created
subscription.operators.coreos.com/nuxeo-operator created
```

Wait for OLM to deploy the Operator. Note that the CSV has an image reference to the `images` namespace, hence (for OpenShift/CRC) the image puller role assignment above. Verify that the Operator Subscription progressed to an *Install Plan* and a *CSV*, and verify the CSV succeeded:

```shell
$ (m)kubectl get sub,csv,ip,pod -n nuxeo-test
NAME                                              PACKAGE         SOURCE         CHANNEL
subscription.operators.coreos.com/nuxeo-operator  nuxeo-operator  nuxeo-operator alpha

NAME                                                               ... PHASE
clusterserviceversion.operators.coreos.com/nuxeo-operator.v0.x.0   ... Succeeded

NAME                                             CSV                    APPROVAL  APPROVED
installplan.operators.coreos.com/install-fzfkv   nuxeo-operator.v0.x.0  Automatic true

NAME                              READY   STATUS    RESTARTS   AGE
nuxeo-operator-5784fbc7cd-dtht4   1/1     Running   0          2m8s
...
```

The key thing to observe is that the `clusterserviceversion` should be in the **Succeeded** phase. Check the logs of the Nuxeo Operator:

```shell
$ (m)kubectl logs nuxeo-operator-5784fbc7cd-dtht4 -n nuxeo-test
{"level":"info","ts":1591727227.9497278,"logger":"cmd","msg":"Operator Version: 0.x.0"}
{"level":"info","ts":1591727227.9499292,"logger":"cmd","msg":"Go Version: go1.14.2"}
{"level":"info","ts":1591727227.9499755,"logger":"cmd","msg":"Go OS/Arch: linux/amd64"}
{"level":"info","ts":1591727227.95001,"logger":"cmd","msg":"Version of operator-sdk: v0.17.1"}
{"level":"info","ts":1591727227.9504607,"logger":"leader","msg":"Trying to become the leader."}
{"level":"info","ts":1591727230.1280968,"logger":"leader","msg":"No pre-existing lock was found."}
{"level":"info","ts":1591727230.1342938,"logger":"leader","msg":"Became the leader."}
...
```

These Nuxeo Operator logs should look familiar to you if you built and ran the operator outside of the cluster manually in the prior section. Next, Create a Nuxeo CR. From here on down, the steps are identical to the manual deployment section above:

```shell
# CRC
$ kubectl apply -f deploy/examples/nuxeo-cr.yaml -n nuxeo-test
# MK8s
$ mkubectl apply -f hack/microk8s/nuxeo-cr-mk8s.yaml -n nuxeo-test
```

Wait for the Nuxeo objects to be created by the Operator:

```shell
# CRC
$ kubectl get nuxeo,deployment,replicaset,pod,route,service,sa -n nuxeo-test
# MK8s
$ mkubectl get nuxeo,deployment,replicaset,pod,ingress,service,sa -n nuxeo-test

# CRC output shown
NAME                       AGE
nuxeo.nuxeo.com/my-nuxeo   14s

NAME                                     READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/my-nuxeo-cluster   1/1     1            1           14s
deployment.extensions/nuxeo-operator     1/1     1            1           2m4s

NAME                                               DESIRED   CURRENT   READY   AGE
replicaset.extensions/my-nuxeo-cluster-f4cbc976b   1         1         1       14s
replicaset.extensions/nuxeo-operator-78b864d584    1         1         1       2m4s

NAME                                   READY   STATUS    RESTARTS   AGE
pod/my-nuxeo-cluster-f4cbc976b-877tc   1/1     Running   0          14s
pod/nuxeo-operator-78b864d584-9tq8m    1/1     Running   0          2m4s
pod/nuxeo-operator-index-fmwbh         1/1     Running   0          5m15s

NAME                                              HOST/PORT                       ...
route.route.openshift.io/my-nuxeo-cluster-route   nuxeo-server.apps-crc.testing   ...

NAME                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   ...
service/my-nuxeo-cluster-service   ClusterIP   172.30.96.168   <none>        80/TCP    ...
service/nuxeo-operator             ClusterIP   172.30.18.240   <none>        50051/TCP ...
service/nuxeo-operator-metrics     ClusterIP   172.30.63.66    <none>        8383/TCP  ...

NAME                            SECRETS   AGE
...
serviceaccount/nuxeo            2         14s
serviceaccount/nuxeo-operator   2         2m4s
```

Access the Nuxeo cluster with your browser and log in to this unlicensed development server with Administrator/Administrator:

```shell
http://nuxeo-server.apps-crc.testing
```

Tear down the Nuxeo cluster:

```shell
$ (m)kubectl delete nuxeo my-nuxeo -n nuxeo-test
```

Allow a moment or two for the cluster to remove the resources. Then:

```shell
# CRC
$ kubectl get nuxeo,deployment,replicaset,pod,route,service,sa,csv,ip,sub -n nuxeo-test
# MK8s
$ mkubectl get nuxeo,deployment,replicaset,pod,ingress,service,sa,csv,ip,sub -n nuxeo-test

# CRC output shown
NAME                                   READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/nuxeo-operator   1/1     1            1           7m33s

NAME                                              DESIRED   CURRENT   READY   AGE
replicaset.extensions/nuxeo-operator-78b864d584   1         1         1       7m33s

NAME                                  READY   STATUS    RESTARTS   AGE
pod/nuxeo-operator-78b864d584-9tq8m   1/1     Running   0          7m33s
pod/nuxeo-operator-index-fmwbh        1/1     Running   0          10m

NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   ...
service/nuxeo-operator-index     ClusterIP   172.30.18.240   <none>        50051/TCP ...
service/nuxeo-operator-metrics   ClusterIP   172.30.63.66    <none>        8383/TCP  ...

NAME                            SECRETS   AGE
...
serviceaccount/nuxeo-operator   2         7m33s

NAME                                                               DISPLAY          ...
clusterserviceversion.operators.coreos.com/nuxeo-operator.v0.x.0   Nuxeo Operator   ...

NAME                                             CSV                     APPROVAL  ...
installplan.operators.coreos.com/install-8lgcq   nuxeo-operator.v0.x.0   Automatic ...

NAME                                               PACKAGE          SOURCE         ...
subscription.operators.coreos.com/nuxeo-operator   nuxeo-operator   nuxeo-operator ...
```
You should be able to observe that all of the Nuxeo *cluster* resources have been removed, and only the Nuxeo Operator and related OLM components remain. Go ahead and remove the namespaces to clean up your environment:

```shell
# CRC
$ kubectl delete namespace nuxeo-test images custom-operators
# MK8s
$ mkubectl delete namespace nuxeo-test images

```

That's a basic demonstration of the capability of the Nuxeo Operator as of the current version.

------

### End to end testing

The e2e testing uses the `operator-sdk local` capability. The e2e testing has a handful of prerequisites:

1. Create the e2e test namespace: `(m)kubectl create namespace operator-test`
2. The e2e test creates a Deployment that references the operator image, and the operator creates a Deployment that references the Nuxeo image. So make sure that the current version of the `nuxeo-operator` image and the 10.10 `nuxeo` image are both present as image streams in the OpenShift `images` namespace, or in the MK8s registry add-on. The steps to accomplish this are documented above in this README.
3. For OpenShift, ensure that the `operator-test` namespace can pull images from the `images` namespace: `oc project operator-test; oc policy add-role-to-group system:image-puller system:serviceaccounts:operator-test --namespace=images`
4. The Make file provides a target `operator-e2e-test` to run the e2e tests. It references the internal OpenShift/MK8s registry using a variable that you can inspect by running Make with the `--just-print` option.
5. For MK8s, make sure you created a cluster role binding as shown in the *microk8s.md* document in the docs folder. Once these prerequisites are satisfied, then run the e2e tests with Make:

```shell
# CRC
$ make operator-e2e-test

# MK8s
$ mkubectl get clusterrolebinding microk8s-e2e
NAME           AGE
microk8s-e2e   1h
$ TARGET_CLUSTER=MICROK8S make operator-e2e-test
```

