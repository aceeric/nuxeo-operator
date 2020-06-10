# Nuxeo Operator

This project is a very early - **0.1.0 at present** - OpenShift/Kubernetes Operator written in Go to manage the state of a Nuxeo cluster. Nuxeo is an open source content management system. (See https://www.nuxeo.com/). The Operator scaffolding was initially generated using the Operator SDK (https://docs.openshift.com/container-platform/4.4/operators/operator_sdk/osdk-getting-started.html/).

Presently, I'm doing this development on a Ubuntu 18.04 desktop and OpenShift Code Ready Containers (https://github.com/code-ready/crc).

### Planned Features

Below is the sequence of capabilities that are planned for this Operator. These are all preliminary and will be tuned as I get further into the project.

#### Version 0.1.0 *(complete)*

This version is really just a POC of the Operator with basic functionality. The goal is to be able to bring up - and reconcile the state of - a basic Nuxeo cluster that optionally supports TLS via an Nginx reverse proxy. The goal is to do this manually, and also via an OLM subscription.

This version creates/reconciles a Deployment, a Service, a Service Account, and an OpenShift Route. (I'm starting with OpenShift but will support Kubernetes in a later version.) It also includes the ability to install Nuxeo Marketplace packages (https://connect.nuxeo.com/nuxeo/site/marketplace) assuming Internet access to the Marketplace by the Operator.

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
| Support running the Operator binary only externally to the cluster to verify the basic Operator functionality | complete |
| Implement a minimal Status field of the Nuxeo CR consisting of the number of active pods  | complete |
| Perform basic testing whereby cluster resources are modified and expected resource reconciliation is performed by the Nuxeo Operator | complete |
| Add the basic elements (CSV, RBACs, bundling, etc.) to support packaging the Operator as a community Operator | complete |
| Deploy and test the Operator from an internal Operator registry in the cluster via OLM subscription | complete |
| Automate the build with *Make*                               | complete |



#### Version 0.2.0

Version 0.2.0 will focus on productionalization by implementing unit testing, and end-to-end testing, and extending the build automation to include these

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Incorporate comprehensive unit testing into the operator build |        |
| Incorporate end-to-end testing                               |        |



#### Version 0.3.0

Version 0.3.0 will focus on Kubernetes support. My belief at present is that the main significant difference will be creating/reconciling a Kubernetes *Ingress* object rather than an OpenShift *Route* object.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Implement the ability to detect whether the Operator is running in a Kubernetes cluster vs. an OpenShift cluster |        |
| Create an *Ingress* resource for access outside of the Kubernetes cluster |        |
| Support a Kubernetes-suitable Dockerfile for the Operator (Operator SDK generates a Dockerfile based on ubi) |        |
| Document Kubernetes testing using MicroK8s (https://microk8s.io/) |        |
| Will need to run a private registry in MicroK8s for equivalence with OpenShift image stream |        |



#### Version 0.4.0

Version 0.4.0 will focus on adding additional functionality to the Operator that I feel is required to make the Operator suitable for consideration by a general audience.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Support a Secret with payload for TLS termination in the Route. Previously, TLS passthrough was the only tested Route/Ingress functionality |        |
| Support a secret for JVM-wide PKI configuration in the Nuxeo Pod - in order to support cases where Nuxeo is running in a PKI-enabled enterprise and is interacting with internal PKI-enabled Corporate micro-services that use an internal corporate CA |        |
| Support installing marketplace packages in disconnected mode if no Internet connection is available in-cluster |        |
| Support readiness and liveness probes for the Nuxeo pods     |        |
| Support explicit definition of nuxeo.conf properties in the Nuxeo CR |        |
| Ability to define *Interactive* nodes and *Worker* nodes separately resulting in two Deployments. The objective is to support compute-intensive back-end processing on a set of nodes having a greater resource share in the cluster then the interactive nodes that serve the Nuxeo GUI |        |



#### Version 0.5.0

Version 0.5.0 will focus on supporting the *Service Binding Operator* to facilitate integration of a Nuxeo Cluster with backing services such as PostgreSQL, Kafka, and ElasticSearch.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Integrate with the Service Binding Operator (https://github.com/redhat-developer/service-binding-operator) to bind Nuxeo to various backing services present in the cluster |        |
| Test with Strimzi (https://strimzi.io/) for Nuxeo Stream support |        |
| Test with Crunchy PostgreSQL (https://www.crunchydata.com/products/crunchy-postgresql-for-kubernetes/) for database support |        |
| Test with Elastic Cloud on Kubernetes (https://github.com/elastic/cloud-on-k8s) for ElasticSearch support |        |
| Support password changes in backing services that trigger rolling updates of the Nuxeo cluster |        |
| Support certificate expiration and replacement for the JVM, and also for backing services that trigger rolling updates of the Nuxeo cluster. An example would be where the Kafka backing service is accessed via TLS, and the Kafka CA and cert expire and are renewed |        |



#### Version 0.6.0

Version 0.6.0 will focus on making the Operator available as a Community Operator

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Gain free access to a full production-grade OpenShift cluster, and a full production-grade Kubernetes cluster to ensure compatibility with those production environments |        |
| Develop and test the elements needed to qualify the Operator for evaluation as a community Operator. Submit the operator for evaluation. Iterate |        |
| Provide `kustomize` examples to illustrate bringing up an exemplar Nuxeo Cluster using kustomize) https://kubernetes.io/docs/tasks/manage-kubernetes-objects/kustomization/) |        |
| Support multi-architecture build. Incorporate lint, gofmt, etc. into the build process |        |
| Make the Operator available as a community Operator (https://github.com/operator-framework/community-operators) |        |



#### Long Term

| Feature                                       | Status |
| --------------------------------------------- | ------ |
| LOE to reach Phase V Operator Maturity Model? |        |
| OperatorHub availability?                     |        |
| Other?...                                     |        |



------

### Testing version 0.1.0 of the Operator with manual deployment

These are the steps you can follow to test version 0.1.0 of the Operator via manual deployment. Following this section, instructions are provided for testing via OLM Subscription.

##### Assumptions

The documentation that follows assumes that you have access to an OpenShift cluster as `kubeadmin`. As stated earlier, I am using Code Ready Containers (CRC) on Ubuntu 18.04 desktop.

The default CRC install assumes a different Linux lineage then Ubuntu and so getting CRC to run successfully on Ubuntu requires manual intervention in the networking configuration as part of the CRC installation process. Presently, for me, I work around this by configuring my etc/hosts with certain entries - shown below - to support testing.

Specifically, note the `nuxeo-server.apps-crc.testing` host name - which maps to the Route generated by the Operator in this testing scenario. (This will be made clear further on down in the README.)

```shell
192.168.130.1   crc.testing
192.168.130.11  api.crc.testing
192.168.130.11  apps-crc.testing
192.168.130.11  foo.apps-crc.testing
192.168.130.11  oauth-openshift.apps-crc.testing
192.168.130.11  default-route-openshift-image-registry.apps-crc.testing
192.168.130.11  nuxeo-server.apps-crc.testing
192.168.130.11  console-openshift-console.apps-crc.testing
```

The IP addresses shown above are hard-coded in the CRC binary for the CRC version shown below, and with the corresponding CRC configuration settings:

```shell
$ ./crc version
crc version: 1.9.0+a68b5e0
OpenShift version: 4.3.10 (embedded in binary)
$ ./crc config view
- skip-check-crc-dnsmasq-file           : true
- skip-check-network-manager-config     : true
- skip-check-network-manager-installed  : true
- skip-check-network-manager-running    : true
```

##### Steps

Ensure you're logged in to the cluster as `kubeadmin`:

```shell
$ oc status
In project default on server https://api.crc.testing:6443
...
$ oc whoami
kube:admin
```

Since the project is still in an early development stage, and is being tested on CRC, using image streams for storing images simplifies testing because these dev images will always be available in-cluster. Create an `images` project to hold various image streams: 

```shell
$ oc new-project images
```

Create a Nuxeo `ImageStream` in the `images` project. These steps use `docker` but `podman` should also work. Also note that I have Docker configured for insecure access to the internal CRC registry:

```shell
$ sudo cat /etc/docker/daemon.json
{
    "insecure-registries" : ["default-route-openshift-image-registry.apps-crc.testing"]
}
```

Create the image stream:

```shell
$ docker pull nuxeo:10.10
$ HOST=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
$ docker tag nuxeo:10.10 $HOST/images/nuxeo:10.10
$ docker login -u kubeadmin -p $(oc whoami -t) $HOST
$ docker push $HOST/images/nuxeo:10.10
```

Deploy the Nuxeo CRD into the cluster:

```shell
$ oc apply -f deploy/crds/nuxeo.com_nuxeos_crd.yaml
customresourcedefinition.apiextensions.k8s.io/nuxeos.nuxeo.com configured
```

Create a `nuxeo` project to test the Operator in, and grant all service accounts in the `nuxeo` project the ability to pull images from the `images` project:
```shell
$ oc new-project nuxeo
$ oc policy add-role-to-group system:image-puller system:serviceaccounts:nuxeo\
  --namespace=images
```

Create a Nuxeo CR for a single-node Nuxeo instance with port 80 access external to the cluster:
```shell
$ oc apply -f deploy/examples/nuxeo-cr.yaml
nuxeo.nuxeo.com/my-nuxeo created
```

Build the Operator binary. This step assumes the current working directory is the directory into which you git cloned this project. Note - this project is Go 1.14. (This build command builds the executable to the same location as the Make file):

```shell
$ go version
go version go1.14.2 linux/amd64
$ go build -o build/_output/bin/nuxeo-operator cmd/manager/main.go
```

Run the Operator outside of the cluster from the command line. To run the operator this way, you provide the  watch namespace as an environment variable, and a command-line option specifying the path of a kube config with credentials for the cluster:

```shell
$ WATCH_NAMESPACE=nuxeo build/_output/bin/nuxeo-operator --kubeconfig=$HOME/.kube/config
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
$ oc get service,sa,deployment,replicaset,route,pod
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

Test access to Nuxeo. Of key importance here is that the host name in the route - `nuxeo-server.apps-crc.testing` - is addressable from your local host:

```shell
$ ping nuxeo-server.apps-crc.testing
PING nuxeo-server.apps-crc.testing (192.168.130.11) 56(84) bytes of data.
64 bytes from api.crc.testing (192.168.130.11): icmp_seq=1 ttl=64 time=0.333 ms
...
$ curl http://nuxeo-server.apps-crc.testing
<META HTTP-EQUIV="refresh" CONTENT="0;URL=/nuxeo">
```

Assuming success so far, then it should be possible to access the Nuxeo Pod from your browser using the same HTTP URL shown in the curl example above. If everything worked you should see the Nuxeo main page:

![](resources/images/nuxeo-ui.jpg)

To review what happened so far:

1. You deployed a Nuxeo CR that specified the `nuxeo-web-ui` Marketplace package
2. You ran the operator from your desktop
3. The Operator saw the Nuxeo CR you generated, and created a Route, a Service, a Service Account, and a Deployment. The Operator passed the `NUXEO_PACKAGES` environment variable from the Nuxeo CR into the Deployment
4. OpenShift reconciled the Deployment into a ReplicaSet and a Pod. The Pod also got the `NUXEO_PACKAGES` environment variable. The Pod runs under the generated `nuxeo` Service Account
5. The Pod started the Nuxeo Container
6. The Nuxeo Container saw the environment variable `NUXEO_PACKAGES=nuxeo-web-ui` and so Nuxeo reached out over the Internet to a hard-coded Marketplace URL, got the package, and installed it into the Nuxeo Container
7. You should now be able to log into this unlicensed development version of Nuxeo as `Administrator/Administrator` (make sure cookies are enabled)

##### **Next, test TLS access using an Nginx sidecar**

Create an *Ningx* ImageStream in the `images` namespace. This step uses the `$HOST` environment variable defined above:

```shell
$ echo $HOST
# yours might be different:
default-route-openshift-image-registry.apps-crc.testing
$ docker pull nginx:latest
$ docker tag nginx:latest $HOST/images/nginx:latest
$ docker push $HOST/images/nginx:latest
latest: digest: sha256:... size: 1362
$ oc get imagestreams -n images
NAME    IMAGE REPOSITORY                                                      TAGS
nginx   default-route-openshift-image-registry.apps-crc.testing/nuxeo/nginx   latest
nuxeo   default-route-openshift-image-registry.apps-crc.testing/nuxeo/nuxeo   10.10
```


Generate a self-signed TLS certificate and key, and a `dhparams` file to use in terminating the TLS connection in the Nginx sidecar:
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

Now, generate the Secret you created above, as well as a ConfigMap provided with the Operator. Both of these resources are used by the Operator to configure Nginx. This assumes you are still in the `tmp` directory:

```shell
$ oc project
Using project "nuxeo" on server "https://api.crc.testing:6443".
$ oc apply -f tls-secret.yaml
secret/tls-secret created
$ oc apply -f ../deploy/examples/tls-cfgmap.yaml
configmap/tls-cfgmap created
$ oc get secret,cm
NAME                              TYPE                                  DATA   AGE
...
secret/tls-secret                 kubernetes.io/tls                     3      84s

NAME                   DATA   AGE
configmap/tls-cfgmap   2      44s

```

Now, create a different Nuxeo CR - this one configures Nuxeo for port 443 access outside the cluster. This CR has a different spec that indicates Nginx is to be used as a reverse proxy:

```shell
$ cd ..
$ oc apply -f deploy/examples/nuxeo-cr-tls.yaml
nuxeo.nuxeo.com/my-nuxeo configured
```

Observe the Operator modifying the Service, and the Deployment. Observe OpenShift regenerating the Pod. Wait for the Pod to stabilize. Note that the Nuxeo Pod now contains two Containers (as indicated by a Ready state of 2/2). The first Container is the Nginx sidecar, and the second Container is Nuxeo:
```shell
$ oc get pod
NAME                                READY   STATUS    RESTARTS   AGE
my-nuxeo-cluster-84b7bc487b-h2rdz   2/2     Running   0          40s
```

You should now be able to access Nuxeo via HTTPS by pasting the Nuxeo URL into your browser:
```shell
https://nuxeo-server.apps-crc.testing/nuxeo
```

Because Nginx is terminating TLS with a self-signed certificate, your browser will display a certificate authenticity warning. Just go ahead and add an exception. That should take you to exactly the same page as you saw before. Check the Nginx container logs to verify that it is acting as the reverse proxy:

```shell
$ oc logs my-nuxeo-cluster-84b7bc487b-h2rdz -c nginx -f
/docker-entrypoint.sh: /docker-entrypoint.d/ is not empty, will attempt to perform configuration
/docker-entrypoint.sh: Looking for shell scripts in /docker-entrypoint.d/
...
{"time_local":"05/Jun/2020:17:20:35 +0000","remote_addr":"10.128.0.1","remote_user":"","request":"GET /nuxeo HTTP/1.1","status": "302","body_bytes_sent":"5","request_time":"0.045","http_referrer":"","http_user_agent":"Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:76.0) Gecko/20100101 Firefox/76.0"}
...
```

Go back to the console that was running the Nuxeo Operator from the command line and press CTRL-C to stop the Operator. You should be returned to the command prompt.

Review the resources created by the Nuxeo Operator. Then, delete the Nuxeo CR and observe that all of those resources are cleaned up automatically by OpenShift via cascading delete, because of the `ownerReferences` on the resources:

```shell
$ oc get nuxeo,deployment,replicaset,pod,route,service,sa
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

$ oc delete nuxeo my-nuxeo
```

Allow a moment or two for OpenShift to remove the resources. Then:

```shell
$ oc get nuxeo,deployment,replicaset,pod,route,service
No resources found in nuxeo namespace.
$ oc get sa/nuxeo
Error from server (NotFound): serviceaccounts "nuxeo" not found
```



------

### Testing version 0.1.0 of the Operator via OLM

These are the steps to test the Nuxeo Operator via OLM. If you didn't execute create an `images` project and put the Nuxeo 10.10 docker image into the project  as an image stream:

```shell
$ oc new-project images
Now using project "images" on server "https://api.crc.testing:6443".
...
$ docker pull nuxeo:10.10
...
Status: Downloaded newer image for nuxeo:10.10
docker.io/library/nuxeo:10.10
$ HOST=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
$ docker tag nuxeo:10.10 $HOST/images/nuxeo:10.10
$ docker login -u kubeadmin -p $(oc whoami -t) $HOST
$ docker push $HOST/images/nuxeo:10.10
...
10.10: digest: sha256:... size: 3259
$ oc get is
NAME    IMAGE REPOSITORY                                                       TAGS  ...
nuxeo   default-route-openshift-image-registry.apps-crc.testing/images/nuxeo   10.10 ...
```

Build and deploy the Nuxeo Operator OLM *Index* - which is what OLM uses to support instantiating the operator via an OLM `subscription`. This section will use the Make file for these tasks. The Make targets that accomplish this are:

1. **operator-build** - build the operator Go binary from Go sources
2. **operator-image-build** - build the operator container image
3. **operator-image-push** - push the operator container image as an image stream to the OpenShift cluster in the `images` namespace
4. **bundle-build** - build the Operator bundle image
7. **index-add** - generate an OLM Index, which references the bundle image
8. **index-push** - push the OLM Index to the OpenShift cluster

Because this is still an early development project, all images are pushed by the Make file as image streams into the OpenShift cluster. The Make file uses the *operator-sdk* command to implement some of the OLM integration. The Make file requires operator-sdk **v0.18.0** so you need to make sure you're running that version. The Make file also uses the **opm** command. You need to build that if it's not already available:

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
$ make operator-build operator-image-build operator-image-push
```

If success , then you will see the Operator image in the `images` namespace, along with the Nuxeo image you pushed previously:

```shell
$ oc get is -n images
NAME             IMAGE REPOSITORY    TAGS    UPDATED
nuxeo            ...nuxeo            10.10   6 minutes ago
nuxeo-operator   ...nuxeo-operator   0.1.0   35 seconds ago
```

Next, generate the OLM components to enable the Operator to be deployed via OLM. The Make file generates the OLM components into a `custom-operators` project, so create that project:

```shell
$ oc new-project custom-operators
Now using project "custom-operators" on server "https://api.crc.testing:6443".
...
```

Now, run the Make file to create the OLM Index that serves the Nuxeo Operator. This step assumes your working directory is project root, and builds on the prior Make step:

```shell
$ make bundle-build index-add index-push
```

If successful, then:

```shell
$ oc get is -n custom-operators
NAME                             IMAGE REPOSITORY                                TAGS  ...
nuxeo-operator-index             default-route-openshift-image-registry.apps...  0.1.0 ...
nuxeo-operator-manifest-bundle   default-route-openshift-image-registry.apps...  0.1.0 ...
```


You can now use the Operator's OLM integration to instantiate a Nuxeo cluster. The summary steps are:

1. Create a new project for the Nuxeo cluster
2. Create a `CatalogSource` in the project to serve the OLM registry components in the project
3. Subscribe the Nuxeo Operator via OLM
4. Wait for the Operator to be running
5. Create a Nuxeo CR
6. Wait for the various Nuxeo objects to be created by the Operator
7. Access Nuxeo from your browser
8. Tear down the Nuxeo cluster
9. Remove the project

These steps are now presented in detail:

Create a project, and grant the project authorization to pull images from the `images` namespace **and** the `custom-operators` namespace:

```shell
$ oc new-project nuxeo-test
Now using project "nuxeo-test" on server "https://api.crc.testing:6443".
...
$ oc policy add-role-to-group system:image-puller system:serviceaccounts:nuxeo-test\
  --namespace=images
# you may get a warning like this - ignore it
Warning: Group 'system:serviceaccounts:nuxeo-test' not found

clusterrole.rbac.authorization.k8s.io/system:image-puller added: "system:serviceaccounts:nuxeo-test"
$ oc policy add-role-to-group system:image-puller system:serviceaccounts:nuxeo-test\
  --namespace=custom-operators
```

Create a `CatalogSource` in the `nuxeo-test` namespace to serve the OLM registry components in the project. Note that the steps below use a SHA reference because if you build the operator index multiple times, the OLM caching behavior won't find newer images if the tag doesn't change. So using a SHA isn't mandatory the first time, but following this process always works. And - note that the catalog source pulls an image from the custom-operators namespace and so the image puller role created above is mandatory:

```shell
$ SHA=$(oc get is nuxeo-operator-index -n custom-operators -o jsonpath='{@.status.tags[0].items[0].image}')
# below is an example - your SHA may differ
$ echo $SHA
sha256:22621b4ea0745b10df4a1e2d985c6b00d74b8a0d16ac42dac4001fcdc3d36d6c
$ cat <<EOF | oc apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: nuxeo-operator
  namespace: nuxeo-test
spec:
  displayName: Nuxeo Operator
  publisher: Test
  sourceType: grpc
  image: image-registry.openshift-image-registry.svc.cluster.local:5000/custom-operators/nuxeo-operator-index@$SHA
EOF
```

Confirm the `CatalogSource ` is operating correctly. It may take a moment for the pod to reach the *Running* state:

```shell
$ oc get catalogsource,pod
NAME             DISPLAY          TYPE   PUBLISHER   AGE
nuxeo-operator   Nuxeo Operator   grpc   Test        4m31s

NAME                   READY   STATUS    RESTARTS   AGE
nuxeo-operator-k4r8p   1/1     Running   0          5m17s

$ oc logs nuxeo-operator-k4r8p
time="2020-06-09T18:04:12Z" level=info msg="serving registry" database=/database/index.db port=50051
```

If the Nuxeo OLM Index Pod was correctly built and is running successfully, it will be hosting a gRPC server serving on port 50051. The logs will be a single line as shown above.

Next, create an Operator Group, and then subscribe to the Nuxeo Operator via OLM:

```shell
$ cat <<EOF | oc apply -f -
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

Wait for OLM to deploy the Operator. Note that the CSV has an image reference to the `images` namespace, hence the image puller role assignment above. Verify that the Operator Subscription progressed to an *Install Plan* and a *CSV*, and verify the CSV succeeded:

```shell
$ oc get sub,csv,ip,pod
NAME                                               PACKAGE          SOURCE           CHANNEL
subscription.operators.coreos.com/nuxeo-operator   nuxeo-operator   nuxeo-operator   alpha

NAME                                                               ... PHASE
clusterserviceversion.operators.coreos.com/nuxeo-operator.v0.1.0   ... Succeeded

NAME                                             CSV                    APPROVAL  APPROVED
installplan.operators.coreos.com/install-fzfkv   nuxeo-operator.v0.1.0  Automatic true

NAME                              READY   STATUS    RESTARTS   AGE
nuxeo-operator-5784fbc7cd-dtht4   1/1     Running   0          2m8s
...
```

The key thing to observe is that the `clusterserviceversion` should be in the **Succeeded** phase. Check the logs of the Nuxeo Operator:

```shell
$ oc logs nuxeo-operator-5784fbc7cd-dtht4
{"level":"info","ts":1591727227.9497278,"logger":"cmd","msg":"Operator Version: 0.1.0"}
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
$ oc apply -f deploy/examples/nuxeo-cr.yaml
```

Wait for the Nuxeo objects to be created by the Operator:

```shell
$ oc get nuxeo,deployment,replicaset,pod,route,service,sa
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
pod/nuxeo-operator-fmwbh               1/1     Running   0          5m15s

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

Access the Nuxeo cluster using  the URL, and log in to this unlicensed development server with Administrator/Administrator:

```shell
http://nuxeo-server.apps-crc.testing
```

Tear down the Nuxeo cluster:

```shell
$ oc delete nuxeo my-nuxeo
```

Allow a moment or two for OpenShift to remove the resources. Then:

```shell
$ $ oc get nuxeo,deployment,replicaset,pod,route,service,sa,csv,ip,sub
NAME                                   READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/nuxeo-operator   1/1     1            1           7m33s

NAME                                              DESIRED   CURRENT   READY   AGE
replicaset.extensions/nuxeo-operator-78b864d584   1         1         1       7m33s

NAME                                  READY   STATUS    RESTARTS   AGE
pod/nuxeo-operator-78b864d584-9tq8m   1/1     Running   0          7m33s
pod/nuxeo-operator-fmwbh              1/1     Running   0          10m

NAME                             TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)   ...
service/nuxeo-operator           ClusterIP   172.30.18.240   <none>        50051/TCP ...
service/nuxeo-operator-metrics   ClusterIP   172.30.63.66    <none>        8383/TCP  ...

NAME                            SECRETS   AGE
serviceaccount/builder          2         13m
serviceaccount/default          2         13m
serviceaccount/deployer         2         13m
serviceaccount/nuxeo-operator   2         7m33s

NAME                                                               DISPLAY          ...
clusterserviceversion.operators.coreos.com/nuxeo-operator.v0.1.0   Nuxeo Operator   ...

NAME                                             CSV                     APPROVAL  ...
installplan.operators.coreos.com/install-8lgcq   nuxeo-operator.v0.1.0   Automatic ...

NAME                                               PACKAGE          SOURCE         ...
subscription.operators.coreos.com/nuxeo-operator   nuxeo-operator   nuxeo-operator ...
```
You should be able to observe that all of the Nuxeo cluster resources have been removed, and only the Nuxeo Operator and related OLM components remain. Go ahead and remove the projects to clean up your environment:

```shell
$ oc delete project nuxeo-test images custom-operators
project.project.openshift.io "nuxeo-test" deleted
project.project.openshift.io "images" deleted
project.project.openshift.io "custom-operators" deleted
```



That's a basic demonstration of the capability of the Nuxeo Operator as of version 0.1.0.

