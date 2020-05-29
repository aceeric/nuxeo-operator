
16-MAY
https://docs.openshift.com/container-platform/4.4/operators/operator_sdk/osdk-getting-started.html

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

hand-modify go.mod to remove abs path and also abs path in main.go imports

```bash
$ cd nuxeo-operator
$ build ./...
$ operator-sdk add api --api-version=nuxeo.com/v1alpha1 --kind=Nuxeo
```

hand edit /home/eace/go/nuxeo-operator/pkg/apis/nuxeo/v1alpha1/nuxeo_types.go
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

Nuxeo Operator Features
- Ability to define interactive nodes and cluster nodes separately resuling in two Deployments
  - requires work queue config per https://doc.nuxeo.com/nxdoc/work-and-workmanager/
- Support Pod Templates for interactive and worker with a hard-coded default configuration
- Runs within a _Nuxeo_ Service (OLM handles this via the CSV)
- OLM also handles RBAC
- Somehow integrates with Svc Binding Op
- Supports optional TLS Nginx sidecar
- Creates a Route
- Supports a Secret with payload for SSL termination in nginx
- Supports a secret for JVM-wide PKI configuration
- Supports custom Nuxeo images
- Allows nuxeo.conf settings
- Implements extensive unit testing
- Implement extensive fake testing
- Implements extensive integration testing with ephemeral OpenShift / Kubernetes cluster
- Supports OpenShift
- Supports Kubernetes
- Supports OLM via CSV, etc.
- Hosted on OperatorHub
- Available as a Community Operator
- Support Nuxeo Connect for installing marketplace packages
- Support installing marketplace packages in disconnected mode
- Supports Make for creating and deploying the Operator Image and OLM assets




Later:

## canonical reconcile logic - per resource
// found := blank
// expected := default from CR
// found = get
// if not found
//   create from new
// else if err
//   err
// else if semantic found != new
//   deep copy new into found
//   update found









https://github.com/nuxeo/nuxeo-apb-catalog/tree/master/nuxeo-apb

$ docker pull nuxeo:10.10
$ docker run --name mynuxeo --rm -ti -p 8080:8080 -e NUXEO_PACKAGES="nuxeo-web-ui" nuxeo:10.10
$ docker tag nuxeo:10.10 nuxeo
$ HOST=$(oc get route default-route -n openshift-image-registry --template='{{ .spec.host }}')
$ docker tag nuxeo:10.10 $HOST/poc-images/nuxeo:10.10
$ docker login -u kubeadmin -p $(oc whoami -t) default-route-openshift-image-registry.apps-crc.testing
$ docker push $HOST/poc-images/nuxeo:10.10

Route
  selector + targetport >
    Service
      selector + port + targetpod >
        Pod
          container
            port

if revproxy specified:
  route listens on 443 and selects 



steps
1 create project poc-rco
2 create project poc-images and give poc-rco permission to pull images from poc-images
3 tag and push nuxeo:10.10 to poc-images/nuxeo:10.10
4 oc apply nuxeo-operator/deploy/crds/nuxeo.com_nuxeos_crd.yaml
5 oc apply nuxeo-operator/deploy/crds/nuxeo.com_v1alpha1_nuxeo_cr.yaml
6 oc apply nuxeo-operator/deploy/crds/tls_secret.yaml
7 oc apply nuxeo-operator/deploy/crds/tls_cfgmap.yaml
8 bring up the operator in an IDE (GoLand)
  needs env vars: WATCH_NAMESPACE=poc-rco
  needs cmdline arg: --kubeconfig=~/.kube/config
9 Run - let it stablilize its reconciliation loop
10 Confirm:
   deployment: my-nuxeo-cluster
   pod:
     container: nuxeo
     container: nginz
   service: my-nuxeo-cluster-service
   route: my-nuxeo-cluster-route












