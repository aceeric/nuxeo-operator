## Go Language Nuxeo Operator (in progress)

### Notes from installing the Operator SDK
(Based on https://docs.openshift.com/container-platform/4.4/operators/operator_sdk/osdk-getting-started.html)

```bash
$ export RELEASE_VERSION=v0.15.0
$ curl -OJL https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu
$ curl -OJL https://github.com/operator-framework/operator-sdk/releases/download/${RELEASE_VERSION}/operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu.asc
$ gpg --verify operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu.asc
gpg: assuming signed data in 'operator-sdk-v0.15.0-x86_64-linux-gnu'
gpg: Signature made Wed 22 Jan 2020 05:20:50 PM EST
gpg:                using RSA key 5E54D124A2DCF2942150F55E66764113082F7CAC
gpg: Can't check signature: No public key

$ gpg --keyserver keys.gnupg.net --recv-key "5E54D124A2DCF2942150F55E66764113082F7CAC"
gpg: key 66764113082F7CAC: public key "Jeff McCormick <jemccorm@redhat.com>" imported
gpg: Total number processed: 1
gpg:               imported: 1

$ chmod +x operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu
$ sudo cp operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu /usr/local/bin/operator-sdk
$ rm operator-sdk-${RELEASE_VERSION}-x86_64-linux-gnu
$ rm operator-sdk-v0.15.0-x86_64-linux-gnu.asc
$ operator-sdk version
operator-sdk version: "v0.15.0", commit: "21a93ca379b887ab2303b0d148a399bf205c3231", go version: "go1.13.5 linux/amd64"
```

```bash
$ cd ~/go
$ operator-sdk new nuxeo-operator --repo ~/go/nuxeo-operator
```

Hand-modify `go.mod `to remove abs path and also abs path in `main.go` imports

```bash
$ cd nuxeo-operator
$ build ./...
$ operator-sdk add api --api-version=nuxeo.com/v1alpha1 --kind=Nuxeo
```

Hand edit `pkg/apis/nuxeo/v1alpha1/nuxeo_types.go`
```bash
# GOROOT IS IMPORTANT
$ export GOROOT=$(go env GOROOT)
$ operator-sdk generate k8s
INFO[0000] Running deepcopy code-generation for Custom Resource group versions: [nuxeo:[v1alpha1], ] 
INFO[0002] Code-generation complete.                    
```
then
```
$ operator-sdk generate crds
```

then
```
$ operator-sdk add controller --api-version=nuxeo.com/v1alpha1 --kind=Nuxeo
```

```
oc apply -f /home/eace/go/nuxeo-operator/deploy/crds/nuxeo.com_nuxeos_crd.yaml
```

Any time `pkg/apis/nuxeo/v1alpha1/nuxeo_types.go` is modified, remember to `generate k8s` and `generate crds`. These steps **require** that the `GOROOT` environment variable is defined as shown above.



### Nuxeo Operator - Planned Features

| Feature                                                      | Status      |
| ------------------------------------------------------------ | ----------- |
| Ability to define Interactive nodes and Worker nodes separately resulting in two Deployments | In progress |
| Support per-deployment work queue config per https://doc.nuxeo.com/nxdoc/work-and-workmanager/ |             |
| Support Pod Templates in the Nuxeo CR for fine-grained configuration of Pods. Use a default configuration, if no Pod Template provided | In progress |
| Run within a _Nuxeo_ Service (OLM handles this via the CSV)  |             |
| Configure RBAC via OLM/CSV                                   |             |
| Integrate with the Service Binding Operator to bind to backing services already present in the OpenShift cluster |             |
| Support an optional Nginx TLS sidecar                        | In progress |
| Create a Route resource for access outside of the OpenShift cluster | In progress |
| Create an Ingress resource for access outside of the Kubernetes cluster |             |
| Support a Secret with payload for TLS termination in Nginx   | In progress |
| Support a secret for JVM-wide PKI configuration in Nuxeo     |             |
| Support custom Nuxeo images, with a default of "nuxeo:nuxeo" if no custom image provided | In progress |
| Support nuxeo.conf settings                                  |             |
| Comprehensive unit testing                                   |             |
| Comprehensive fake testing for simulating in-cluster tests   |             |
| Comprehensive integration testing with ephemeral OpenShift / Kubernetes cluster |             |
| Supports OpenShift                                           | In progress |
| Supports Kubernetes                                          |             |
| Supports OLM via CSV, etc.                                   |             |
| Available as an OperatorHub Operator                         |             |
| Available as a Community Operator                            |             |
| Support Nuxeo Connect for installing marketplace packages with an Internet connection |             |
| Support installing marketplace packages in disconnected mode if no Internet connection |             |
| Support GNU Make for creating and deploying the Operator Image and packaging / submitting OLM assets |             |
| Support readiness probes and liveness probes                 |             |
| Support password changes in backing services                 |             |
| Support certificate expiration and addition of new certificate |             |
| Next...                                                      |             |
|                                                              |             |
|                                                              |             |



### Nuxeo APB for reference

https://github.com/nuxeo/nuxeo-apb-catalog/tree/master/nuxeo-apb



### Adding a Nuxeo 10.10 Image to the 'poc-images' namespace

```bash
$ docker pull nuxeo:10.10
$ docker run --name mynuxeo --rm -ti -p 8080:8080 -e NUXEO_PACKAGES="nuxeo-web-ui" nuxeo:10.10
$ HOST=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
$ docker tag nuxeo:10.10 $HOST/poc-images/nuxeo:10.10
$ docker login -u kubeadmin -p $(oc whoami -t) default-route-openshift-image-registry.apps-crc.testing
$ docker push $HOST/poc-images/nuxeo:10.10
```



### Steps to test the operator (running out of cluster in the IDE)

1. Create project poc-rco
2. Create project poc-images and give poc-rco permission to pull images from poc-images
3. Tag and push nuxeo:10.10 to poc-images/nuxeo:10.10
4. oc apply nuxeo-operator/deploy/crds/nuxeo.com_nuxeos_crd.yaml
5. oc apply nuxeo-operator/deploy/crds/nuxeo.com_v1alpha1_nuxeo_cr.yaml
6. oc apply nuxeo-operator/deploy/crds/tls_secret.yaml
7. oc apply nuxeo-operator/deploy/crds/tls_cfgmap.yaml
8. bring up the operator in an IDE (GoLand)
   1. needs env vars: WATCH_NAMESPACE=poc-rco
   2. needs cmdline arg: --kubeconfig=<location of your .kube/config>
9. Run the operator - let it stabilize its reconciliation loop
10. Stop the operator
11. Confirm:
    1. deployment: my-nuxeo-cluster
    2. pod:
       1. container: nuxeo
       2. container: nginx
    3. service: my-nuxeo-cluster-service
    4. route: my-nuxeo-cluster-route

### Next Steps

The work done so far only crudely creates cluster resources without the intent of connecting them all. It's just a mechanical exercise in resource creation. Next, I need to actually wire up all the connecting pieces:

​	Route > Service > Pod

... then verify it is possible to access Nuxeo outside the cluster by adding the route host name into /etc/hosts and accessing that host name from the browser using SSL. Confirm the traffic goes through Nginx to Nuxeo.

