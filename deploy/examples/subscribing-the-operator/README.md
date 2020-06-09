### Subscribing the Nuxeo Operator - IN PROGRESS

TODO:

1) use images namespace for : nuxeo, nginx, and nuxeo-operator

2) use custom-operators namespace for: nuxeo-operator-manifest-bundle and nuxeo-operator-index because OLM can get at those

3) Add 0.2.0 to 0.1.0 and dec all subsequent versions





IN PROGRESS FIX >>>>>>>>



This README assumes that you have followed the steps documented in the project Makefile for building and installing the Nuxeo Operator with OLM support into a test cluster from a fresh clone of this GitHub repo. Now, you can use the Operator's OLM integration to instantiate a Nuxeo cluster.

The steps are:

1. Create a project
2. Create a `CatalogSource` in the project to serve the OLM registry components in the project
3. Subscribe to the Operator via OLM, and wait for OLM to deploy the Operator
4. Wait for the Operator to be running
5. Create a Nuxeo CR
6. Wait for the Nuxeo objects to be created by the Operator
7. Access Nuxeo using  the URL
8. Tear down the Nuxeo cluster
9. Remove the project



These steps are now presented in detail:

Create a project

```shell
$ oc new-project nuxeo-test
namespace/nuxeo-test created
```

Create a `CatalogSource` in the `nuxeo-test` project to serve the OLM registry components in the project:

```shell
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
  image: image-registry.openshift-image-registry.svc.cluster.local:5000/custom-operators/nuxeo-operator-index:0.1.0
EOF
```

Confirm the `CatalogSource ` is operating correctly:

```she
$ oc get catalogsource
NAME             DISPLAY          TYPE   PUBLISHER   AGE
nuxeo-operator   Nuxeo Operator   grpc   Test        4m31s
$ oc get pod
NAME                   READY   STATUS    RESTARTS   AGE
nuxeo-operator-k4r8p   1/1     Running   0          5m17s
$ oc logs nuxeo-operator-k4r8p
time="2020-06-09T18:04:12Z" level=info msg="serving registry" database=/database/index.db port=50051
```

TODO:

oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:nuxeo-test:nuxeo-operator



ALSO CSV NEEDS TO REF IMAGE IN REGISTRY



The key thing to observe is that the logs are a one-liner, serving on port 50051.

Subscribe to the Operator via OLM, and wait for OLM to deploy the Operator

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
EOF

$ cat <<EOF | oc apply -f -
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
```



Verify that the Operator Subscription progressed to an InstallPlan and a CSV and the CSV succeeded:

```shell
$ oc get sub,csv,ip,pod
NAME                                               PACKAGE          SOURCE           CHANNEL
subscription.operators.coreos.com/nuxeo-operator   nuxeo-operator   nuxeo-operator   alpha

NAME                                                               ... PHASE
clusterserviceversion.operators.coreos.com/nuxeo-operator.v0.1.0   ... Succeeded

NAME                                             CSV                     APPROVAL    APPROVED
installplan.operators.coreos.com/install-fzfkv   nuxeo-operator.v0.1.0   Automatic   true

NAME                              READY   STATUS    RESTARTS   AGE
nuxeo-operator-5784fbc7cd-dtht4   1/1     Running   0          2m8s
...
```

And...

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

These Nuxeo Operator logs should look familiar to you if you built and ran the operator outside of the cluster manually.

Next, Create a Nuxeo CR

```shell
oc apply -f deploy/examples/nuxeo-cr.yaml

WORK-AROUND - CRC ISSUE??
oc adm policy add-cluster-role-to-user cluster-admin system:serviceaccount:nuxeo-test:nuxeo


```

Wait for the Nuxeo objects to be created by the Operator

```shell
etc
```

Access Nuxeo using  the URL

```shell
http://nuxeo-server.apps-crc.testing
```

Tear down the Nuxeo cluster

```shell
as before
```

Remove the project

```shell
as before
```
