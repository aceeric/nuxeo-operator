# Nuxeo Operator

This project is a very early - **0.1.0 at present** - OpenShift/Kubernetes Operator written in Go to manage the state of a Nuxeo cluster. Nuxeo is an open source content management system. (See https://www.nuxeo.com/). The Operator scaffolding was initially generated using the Operator SDK (https://docs.openshift.com/container-platform/4.4/operators/operator_sdk/osdk-getting-started.html/).

Presently, I'm doing this development on a Ubuntu 18.04 desktop and OpenShift Code Ready Containers (https://github.com/code-ready/crc).

### Planned Features

Below is the sequence of capabilities that are planned for this Operator. These are all preliminary and will be tuned as I get further into the project.

#### Version 0.1.0 *(current)*

This version is really just a POC of the Operator with basic functionality. The goal is to be able to bring up - and reconcile the state of - a basic Nuxeo cluster that optionally supports TLS via an Nginx reverse proxy.

This version creates/reconciles a Deployment, a Service, and an OpenShift Route. (I'm starting with OpenShift but will support Kubernetes in a later version.) It also includes the ability to install Nuxeo Marketplace packages (https://connect.nuxeo.com/nuxeo/site/marketplace) assuming Internet access to the Marketplace by the Operator.

| Feature                                                      | Status      |
| ------------------------------------------------------------ | ----------- |
| Define an initial Nuxeo Custom Resource Definition (CRD) that supports the feature set for this increment of functionality | In progress |
| Generate and reconcile a Deployment from a Nuxeo Custom Resource (CR) to represent the desired state of a Nuxeo cluster | In progress |
| Run Nuxeo only in development mode with only embedded services | In progress |
| Support *Pod Templates* in the Nuxeo CR for fine-grained configuration of Pods. Use a default Template if no Pod Template is provided | In progress |
| Support an optional Nginx TLS reverse proxy sidecar container in each Nuxeo Pod to support TLS. Test using 443 outside the cluster, Nginx listening on 8443, and forwarding to Nuxeo on 8080. Model the sidecar configuration from the Nuxeo APB catalog (https://github.com/nuxeo/nuxeo-apb-catalog/blob/master/nuxeo-apb/files/nginx.conf) | In progress |
| Create and reconcile a Route resource for access outside of the OpenShift cluster. Support only TLS passthrough at this time | In progress |
| Create and reconcile a Service resource for the Route to use, and for potential use within the cluster. The service will communicate with Nuxeo on 8080, or Nginx on 8443 | In progress |
| Create all resources that originate from a Nuxeo CR with `ownerReferences` that reference the Nuxeo CR - so that deletion of the Nuxeo CR will result in recursive removal of all generated resources for proper clean up | In progress |
| Support custom Nuxeo images, with a default of `nuxeo:latest` if no custom image is provided in the Nuxeo CR | In progress |
| Support running the Operator binary only externally to the cluster to verify the basic Operator functionality | In progress |
| Perform basic testing whereby cluster resources are modified and expected resource reconciliation is performed by the Nuxeo Operator | In progress |



#### Version 0.2.0

Version 0.2.0 will focus on supporting the Operator Lifecycle Manager (OLM) framework, and running the Operator in-cluster.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Add the necessary elements (CSV, RBACs, bundling, etc.) to support packaging the Operator as a community Operator |        |
| Deploy and test the Operator from a private internal Operator registry in the cluster via OLM subscription |        |



#### Version 0.3.0

Version 0.3.0 will focus on productionalization by implementing unit testing, and end-to-end testing, and automating the build process.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Incorporate comprehensive unit testing into the operator build |        |
| Incorporate end-to-end testing                               |        |
| Automate the build with *Make*                               |        |



#### Version 0.4.0

Version 0.4.0 will focus on Kubernetes support. My belief at present is that the only significant difference will be creating/reconciling a Kubernetes *Ingress* object rather than an OpenShift *Route* object.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Implement the ability to detect whether the Operator is running in a Kubernetes cluster vs. an OpenShift cluster |        |
| Create an *Ingress* resource for access outside of the Kubernetes cluster |        |
| Kubernetes testing using MicroK8s (https://microk8s.io/)     |        |



#### Version 0.5.0

Version 0.5.0 will focus on adding additional functionality to the Operator that I feel is required to make the Operator suitable for consideration by a general audience.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Support a Secret with payload for TLS termination in the Route. Previously, TLS passthrough was the only tested Route/Ingress functionality |        |
| Support a secret for JVM-wide PKI configuration in the Nuxeo Pod - in order to support cases where Nuxeo is running in a PKI-enabled enterprise and is interacting with internal PKI-enabled Corporate micro-services that use a private corporate CA |        |
| Support installing marketplace packages in disconnected mode if no Internet connection is available in-cluster |        |
| Support readiness and liveness probes for the Nuxeo pods     |        |
| Support explicit definition of nuxeo.conf properties in the Nuxeo CR |        |
| Ability to define *Interactive* nodes and *Worker* nodes separately resulting in two Deployments. The objective is to support compute-intensive back-end processing on a set of nodes having a greater resource share in the cluster then the interactive nodes that serve the Nuxeo GUI |        |



#### Version 0.6.0

Version 0.6.0 will focus on supporting the Service Binding Operator to facilitate integration of a Nuxeo Cluster with backing services such as PostgreSQL, Kafka, and ElasticSearch.

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Integrate with the Service Binding Operator (https://github.com/redhat-developer/service-binding-operator) to bind Nuxeo to various backing services present in the cluster |        |
| Test with Strimzi (https://strimzi.io/) for Nuxeo Stream support |        |
| Test with Crunchy PostgreSQL (https://www.crunchydata.com/products/crunchy-postgresql-for-kubernetes/) for database support |        |
| Test with Elastic Cloud on Kubernetes (https://github.com/elastic/cloud-on-k8s) for ElasticSearch support |        |
| Support password changes in backing services that trigger rolling updates of the Nuxeo cluster |        |
| Support certificate expiration and replacement for the JVM, and also for backing services that trigger rolling updates of the Nuxeo cluster. An example would be where the Kafka backing service is accessed via TLS, and the Kafka CA and cert expire and are renewed. |        |



#### Version 0.7.0

Version 0.7.0 will focus on making the Operator available as a Community Operator

| Feature                                                      | Status |
| ------------------------------------------------------------ | ------ |
| Gain free access to a full production-grade OpenShift cluster, and a full production-grade Kubernetes cluster |        |
| Develop and test the elements needed to qualify the Operator for evaluation as a community Operator. Submit the operator for evaluation. Iterate |        |
| Make the Operator available as a community Operator (https://github.com/operator-framework/community-operators) |        |



#### Testing version 0.1.0 (current) of the Operator

These are the steps I follow to test version 0.1.0 of the Operator.

##### Assumptions

The documentation that follows assumes that you have access to an OpenShift cluster as `kubeadmin`. As stated earlier, I am using Code Ready Containers (CRC) on Ubuntu 18.04 desktop.

The default CRC install assumes a different Linux lineage then Ubuntu and so getting CRC to run successfully on Ubuntu requires manual intervention in the networking configuration as part of the CRC installation process. Presently, for me, I work around this by configuring my etc/hosts with certain entries - shown below - to support this testing.

Specifically, note the `nuxeo-server.apps-crc.testing` host name - which maps to the Route generated by the Operator. (This will be made clear further on down in the README.)

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

Ensure you're logged in to the cluster as `kubeadmin`

```shell
$ oc status
In project default on server https://api.crc.testing:6443

svc/openshift - kubernetes.default.svc.cluster.local
svc/kubernetes - 172.30.0.1:443 -> 6443

View details with 'oc describe <resource>/<name>' or list everything with 'oc get all'.
$ oc whoami
kube:admin
```

Create a Nuxeo project in the cluster

```shell
$ oc new-project nuxeo
```

Build the Operator binary. This step assumes you git cloned this project to `$HOME/go/nuxeo-operator` and that is the current working directory.

```shell
$ go build -o bin/nuxeo-operator cmd/manager/main.go
```

Create a Nuxeo `ImageStream` in the Nuxeo project. These steps use `docker` but `podman` should also work. Also note that I have Docker configured for insecure access to the internal CRC registry:

```shell
$ sudo cat /etc/docker/daemon.json
{
    "insecure-registries" : ["default-route-openshift-image-registry.apps-crc.testing"]
}
```

So then:

```shell
$ docker pull nuxeo:10.10
$ HOST=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
$ docker tag nuxeo:10.10 $HOST/nuxeo/nuxeo:10.10
$ docker login -u kubeadmin -p $(oc whoami -t) $HOST
$ docker push $HOST/nuxeo/nuxeo:10.10
```

Deploy the Nuxeo CRD into the cluster

```shell
$ oc apply -f deploy/crds/nuxeo.com_nuxeos_crd.yaml
customresourcedefinition.apiextensions.k8s.io/nuxeos.nuxeo.com configured
$ oc get nuxeos
No resources found in nuxeo namespace.
```

Create a test Nuxeo CR for port 80 access external to the cluster
```shell
$ oc apply -f deploy/examples/nuxeo-cr.yaml
nuxeo.nuxeo.com/my-nuxeo created
```

The operator generates a Nuxeo Deployment that runs under a Service Account "nuxeo". In the future, this Service Account will be created by OLM, but for now, create it manually:

```shell
$ oc apply -f deploy/examples/nuxeo-service-account.yaml
```

Run the Operator outside of the cluster from the command line. To run the operator this way, you provide the  watch namespace as an environment variable, and a command-line option specifying the path of a kube config with credentials for the cluster:

```shell
$ WATCH_NAMESPACE=nuxeo ./bin/nuxeo-operator --kubeconfig=$HOME/.kube/config
```

You should get output like the following:

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
$ oc get service,deployment,replicaset,route,pod
NAME                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S) ...
service/my-nuxeo-cluster-service   ClusterIP   172.30.80.128   <none>        80/TCP  ...

NAME                                     READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/my-nuxeo-cluster   1/1     1            1           4m43s

NAME                                                DESIRED   CURRENT   READY   AGE
replicaset.extensions/my-nuxeo-cluster-6c4b7466dc   1         1         1       4m43s

NAME                                              HOST/PORT                       ...
route.route.openshift.io/my-nuxeo-cluster-route   nuxeo-server.apps-crc.testing   ...

NAME                                    READY   STATUS    RESTARTS   AGE
pod/my-nuxeo-cluster-6c4b7466dc-crw2q   1/1     Running   0          37s
```

Test access via a browser. Of key importance here is that you have to be sure that the host name in the route - `nuxeo-server.apps-crc.testing` - is addressable from your local host:

```shell
$ ping nuxeo-server.apps-crc.testing
PING nuxeo-server.apps-crc.testing (192.168.130.11) 56(84) bytes of data.
64 bytes from api.crc.testing (192.168.130.11): icmp_seq=1 ttl=64 time=0.333 ms
...
$ curl http://nuxeo-server.apps-crc.testing
<META HTTP-EQUIV="refresh" CONTENT="0;URL=/nuxeo">
```

Assuming success, then it should be possible to access the Nuxeo Pod from your browser using the same HTTP URL shown in the curl example above. If everything worked you should see the Nuxeo main page:

![](resources/images/nuxeo-ui.jpg)

To review what happened so far:

1. You deployed a Nuxeo CR that specified the `nuxeo-web-ui` Marketplace package
2. You ran the operator from your desktop
3. The Operator saw the Nuxeo CR you had generated, and created a Route, a Service, and a Deployment. The Operator passed the `NUXEO_PACKAGES` environment variable from the Nuxeo CR into the Deployment
4. OpenShift reconciled the Deployment into a ReplicaSet and a Pod. The Pod also got the `NUXEO_PACKAGES` environment variable
5. The Pod started the Nuxeo Container
6. The Nuxeo Container saw the environment variable `NUXEO_PACKAGES=nuxeo-web-ui` and so Nuxeo reached out over the Internet to a hard-coded Marketplace URL, got the package, and installed it into the Nuxeo Container
7. You should now be able to log into this unlicensed development version of Nuxeo as `Administrator/Administrator` (make sure cookies are enabled)



**Next, test TLS access using an Nginx sidecar**

Create an Ningx ImageStream in the Nuxeo project (uses the `$HOST` environment variable defined above)

```shell
$ docker pull nginx:latest
$ docker tag nginx:latest $HOST/nuxeo/nginx:latest
$ docker push $HOST/nuxeo/nginx:latest
latest: digest: sha256:... size: 1362
$ oc get imagestreams
NAME    IMAGE REPOSITORY                                                      TAGS
nginx   default-route-openshift-image-registry.apps-crc.testing/nuxeo/nginx   latest
nuxeo   default-route-openshift-image-registry.apps-crc.testing/nuxeo/nuxeo   10.10
```


Generate a self-signed TLS certificate and key, and a `dhparams` file to use in terminating the TLS connection in the Nginx sidecar
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

Delete the nuxeo CR and observe that owned resources got cleaned up:

```shell
$ oc get nuxeo,deployment,replicaset,pod,route,service
NAME                       AGE
nuxeo.nuxeo.com/my-nuxeo   106m

NAME                                     READY   UP-TO-DATE   AVAILABLE   AGE
deployment.extensions/my-nuxeo-cluster   1/1     1            1           103m

NAME                                                DESIRED   CURRENT   READY   AGE
replicaset.extensions/my-nuxeo-cluster-6c4b7466dc   0         0         0       103m
replicaset.extensions/my-nuxeo-cluster-84b7bc487b   1         1         1       40m

NAME                                    READY   STATUS    RESTARTS   AGE
pod/my-nuxeo-cluster-84b7bc487b-h2rdz   2/2     Running   0          21m

NAME                                              HOST/PORT                     ...
route.route.openshift.io/my-nuxeo-cluster-route   nuxeo-server.apps-crc.testing ...

NAME                               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S) ...
service/my-nuxeo-cluster-service   ClusterIP   172.30.80.128   <none>        443/TCP ...

$ oc delete nuxeo my-nuxeo
```

Allow a moment or two for OpenShift to remove the resources. Then:

```shell
$ oc get nuxeo,deployment,replicaset,pod,route,service
No resources found in nuxeo namespace.
```



That's a basic demonstration of the capability of the Nuxeo Operator as of version 0.1.0.

