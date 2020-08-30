# Testing off-line installation of Marketplace Packages

This README documents the steps to configure the Nuxeo Operator to install Marketplace Packages into Nuxeo when the Kubernetes cluster does not have Internet connectivity to the Nuxeo Marketplace. This is referred to as *off-line package installation*.

This example uses a ConfigMap to store the package ZIP files. This approach is therefore subject to the Kubernetes limitations on the size of ConfigMap/Secret data elements. According to https://kubernetesbyexample.com/secrets/: "A per-secret size limit of 1MB exists". This also applies to config maps.

As of the current operator version, the Nuxeo Docker entrypoint script does not readily support mounting a persistent volume into the package installation directory. So for the present, only ConfigMaps and Secrets are supported. Hopefully, a change will be made to the Nuxeo Docker entrypoint that will facilitate using PVs and PVCs. If and when that happens, the Operator will be updated to support this.

Begin by creating a workspace on your local filesystem:

```shell
$ mkdir offline
$ cd offline
```

This test will use the *Nuxeo Sample* package and the *Nuxeo Tree Snapshot* package (because they are smaller than 1MB!) Download the packages: 

```shell
$ curl https://connect.nuxeo.com/nuxeo/site/marketplace/package/nuxeo-sample/download?version=2.5.3 -o nuxeo-sample-2.5.3.zip
$ curl https://connect.nuxeo.com/nuxeo/site/marketplace/package/nuxeo-tree-snapshot/download?version=1.2.3 -o nuxeo-tree-snapshot-1.2.3.zip
```

Create ConfigMaps to hold the ZIP files:

```shell
$ oc create configmap nuxeo-sample-marketplace-package --from-file=nuxeo-sample-2.5.3.zip
$ oc create configmap nuxeo-tree-snapshot-marketplace-package --from-file=nuxeo-tree-snapshot-1.2.3.zip
```

With the Nuxeo Operator running, create a Nuxeo CR that references the ConfigMaps:

```shell
cat <<'EOF' | oc apply -f -
apiVersion: appzygy.net/v1alpha1
kind: Nuxeo
metadata:
  name: my-nuxeo
spec:
  nuxeoImage: image-registry.openshift-image-registry.svc.cluster.local:5000/images/nuxeo:10.10
  access:
    hostname: nuxeo-server.apps-crc.testing
  nodeSets:
  - name: cluster
    replicas: 1
    interactive: true
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
      # this is the congiguration that projects the ZIP files into the Nuxeo container
      offlinePackages:
      - packageName: nuxeo-sample-2.5.3.zip
        valueFrom:
          configMap:
            name: nuxeo-sample-marketplace-package
      - packageName: nuxeo-tree-snapshot-1.2.3.zip
        valueFrom:
          configMap:
            name: nuxeo-tree-snapshot-marketplace-package
EOF
```

Observe the package installation activity in the Nuxeo logs:
```shell
$ $ oc logs my-nuxeo-cluster-68d56bc68f-2x8x2
/docker-entrypoint.sh: installing Nuxeo package /docker-entrypoint-initnuxeo.d/nuxeo-sample-2.5.3.zip
...
Added /docker-entrypoint-initnuxeo.d/nuxeo-sample-2.5.3.zip
...
Installing nuxeo-sample-2.5.3
...
/docker-entrypoint.sh: installing Nuxeo package /docker-entrypoint-initnuxeo.d/nuxeo-tree-snapshot-1.2.3.zip
...
Added /docker-entrypoint-initnuxeo.d/nuxeo-tree-snapshot-1.2.3.zip
...
Optional dependencies [nuxeo-jsf-ui] will be ignored for 'nuxeo-tree-snapshot-1.2.3'.
...
Installing nuxeo-tree-snapshot-1.2.3
...
```

Observe how the Operator configured the Pod:
```shell
$ oc describe pod my-nuxeo-cluster-68d56bc68f-2x8x2
...
Containers:
  nuxeo:
    ...
    Mounts:
      /docker-entrypoint-initnuxeo.d/nuxeo-sample-2.5.3.zip...
      /docker-entrypoint-initnuxeo.d/nuxeo-tree-snapshot-1.2.3.zip...
    ...
Volumes:
  offline-package-0:
    Type:      ConfigMap (a volume populated by a ConfigMap)
    Name:      nuxeo-sample-marketplace-package
    Optional:  false
  offline-package-1:
    Type:      ConfigMap (a volume populated by a ConfigMap)
    Name:      nuxeo-tree-snapshot-marketplace-package
    Optional:  false
...
```

List the marketplace ZIP files that were projected into the container, resulting in the installation:
```shell
$ oc exec my-nuxeo-cluster-68d56bc68f-2x8x2 -- ls -l /docker-entrypoint-initnuxeo.d
total 48
-rw-r--r--. 1 root 1000590000 10496 Jun 29 18:04 nuxeo-sample-2.5.3.zip
-rw-r--r--. 1 root 1000590000 33424 Jun 29 18:04 nuxeo-tree-snapshot-1.2.3.zip
```
