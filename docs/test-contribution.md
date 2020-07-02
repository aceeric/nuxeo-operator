# Testing Custom Contributions

This README documents the steps to configure Nuxeo with custom contributions. This supports the ability to configure Nuxeo Pods to start up with different configurations. The core use case is configuring Nuxeo processing nodes differently from nodes serving HTTP(S) as discussed here: https://doc.nuxeo.com/nxdoc/work-and-workmanager/#configuring-dedicated-nodes-in-a-nuxeo-cluster

To configure a Nuxeo CR `NodeSet` with a specific configuration, the configurer follows these steps:

1. Configure a Kubernetes storage resource to hold the contribution elements
2. Configure the Nuxeo CR to reference the storage resource

Then, the Nuxeo Operator reconciles the Nuxeo Pods, adding the contributions. When the Nuxeo server starts, it includes the contribution.

A custom contribution is modeled by the operator as a directory that is mounted into the Nuxeo container at `/etc/nuxeo/nuxeo-operator-config`. Each contribution defined in a storage resource is mounted into its own sub-directory, and added to the Nuxeo templates list with an absolute path. For each contribution, you are required to configure a `nuxeo.defaults`, and one or more contribution files. For example, if a contribution `my-contrib` is defined with a `nuxeo.defaults` and a contribution `my-config.xml`, then the operator creates the following structure:

```shell
etc
└── nuxeo
    └── nuxeo-operator-config
        └── my-contrib
            ├── nuxeo.defaults
            └── nxserver
                └── config
                    └── my-config.xml
```

And then the operator sets `NUXEO_TEMPLATES` to include `/etc/nuxeo/nuxeo-operator-config/my-contrib`. As a result, when the Nuxeo server starts, it loads the contribution, resulting in the Nuxeo server directory `/opt/nuxeo/server/nxserver/config` containing the contribution from the storage resource.

The examples below assume you have a namespace in the cluster to test in, and that the namespace is the current context, and that the Nuxeo Operator is running/watching the namespace.

## Example 1:

Create a Kubernetes storage resource to hold the contribution content.  Any Kubernetes Volume Source can be used to hold the contribution. This first example uses a ConfigMap. A second example below uses a PV/PVC.

This example creates a ConfigMap with a custom Nuxeo Automation Extension as documented here: https://doc.nuxeo.com/nxdoc/automation-chain/. This extension will provide a new automation endpoint - *Document.Query.Extension*. This makes it easy to verify whether the contribution was correctly incorporated into the Nuxeo instance. 

### Create the ConfigMap

The ConfigMap specifies all of the files comprising the contribution:

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: custom-automation-chain-configmap
data:
  nuxeo.defaults: |
    custom-automation-chain.target=.
    custom.automation.chain.property=123
  custom-automation-chain-config.xml:
    <?xml version="1.0"?>
    <component name="custom.automation.chain">
      <extension target="org.nuxeo.ecm.core.operation.OperationServiceComponent"
                 point="chains">
        <chain id="Document.Query.Extension">
          <operation id="Document.Query">
            <param type="string" name="query">SELECT * FROM Document</param>
            <param type="integer" name="pageSize">1</param>
          </operation>
        </chain>
      </extension>
    </component>
EOF
$ kubectl get cm
NAME                                      DATA   AGE
...
custom-automation-chain-configmap         2      4s
...
```

Regarding the ConfigMap above:

You can see that the `data` section has two keys: `nuxeo.defaults` and `custom-automation-chain-config.xml`. The contents of these keys  are projected as files into the Nuxeo Pod as described above. To reiterate: the `nuxeo.defaults` file goes into the root of the contribution subdirectory and everything else goes into the `./nxserver/config` sub-directory under that. E.g.:

```shell
etc
└── nuxeo
    └── nuxeo-operator-config
        └── custom-automation-chain ## where did this come from? explained below...
            ├── nuxeo.defaults
            └── nxserver
                └── config
                    └── custom-automation-chain-config.xml
```

### Create a Nuxeo CR

This CR references the contribution using the `contribs` element:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: nuxeo.com/v1alpha1
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
    livenessProbe:
      exec:
        command:
          - "true"
    readinessProbe:
      exec:
        command:
          - "true"
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
    contribs:
    - volumeSource:
        configMap:
          name: custom-automation-chain-configmap
      templates:
      - custom-automation-chain
EOF
```

Above, you can see that the *contribs* element specifies two things. First, the Kubernetes `VolumeSource` that contains the contribution items. In this case, a ConfigMap. Second, the *templates* element establishes a name for the contribution. It results in the creation of a directory `/etc/nuxeo/nuxeo-operator-config/custom-automation-chain`. And it causes the Operator  to configure the `NUXEO_TEMPLATES` environment var variable with an absolute path reference to that directory.

For ConfigMap and Secret volume sources, only one template name is allowed per Volume Source. Though multiple *contribs* entries are allowed. For PV/PVC contributions, multiple templates are allowed but only one `contribs` entry is allowed. This will be described later in Example Two.

***NOTE:*** An important point regarding the contribution is: the **minimum** content that the configurer must specify for the `nuxeo.defaults` key is:

```shell
custom-automation-chain.target=.
```

Where the token `custom-automation-chain` matches the name of the `contribs.templates` item in the Nuxeo CR. This is used by the Nuxeo loader. Let's say you chose to call the template `foo` in the Nuxeo CR, rather than `custom-automation-chain`. Then the first line of your `nuxeo.defaults` would be:

```shell
foo.target=.
```

### Verify

With the Nuxeo Operator running, examine the Nuxeo Pod startup logs (only relevant content is shown):

```shell
$ kubectl logs my-nuxeo-cluster-7f59f59555-whqrq
...
Include template: /etc/nuxeo/nuxeo-operator-config/custom-automation-chain
...
======================================================================
= Component Loading Status: Pending: 0 / Missing: 0 / Unstarted: 0 / Total: 500
======================================================================
...
```

Inspect the directory structure created by the operator:

```shell
$ kubectl exec my-nuxeo-cluster-7f59f59555-whqrq -- ls -LR /etc/nuxeo/nuxeo-operator-config
/etc/nuxeo/nuxeo-operator-config:
custom-automation-chain

/etc/nuxeo/nuxeo-operator-config/custom-automation-chain:
nuxeo.defaults
nxserver

/etc/nuxeo/nuxeo-operator-config/custom-automation-chain/nxserver:
config

/etc/nuxeo/nuxeo-operator-config/custom-automation-chain/nxserver/config:
custom-automation-chain-config.xml
```

Invoke the newly contributed Automation Chain *Document.Query.Extension*. Take note of the `pageSize` that is returned. It matches the same property from the contribution. This example uses `json_pp` as the JSON formatter.

Here's the curl command:

```shell
$ curl -sH 'Content-Type:application/json+nxrequest' -X POST\
>  -d '{"params":{},"context":{}}'\
>  -u Administrator:Administrator\
>  http://nuxeo-server.apps-crc.testing/nuxeo/api/v1/automation/Document.Query.Extension\
>  | json_pp
{
   ... (omitted for brevity)
   "entries" : [
      {
         "uid" : "a8c3ba11-0d7a-413a-acf1-7fc1ab0c720e",
         "isCheckedOut" : true,
         "state" : "undefined",
         "facets" : [
            "HiddenInNavigation"
         ],
         "path" : "/management/administrative-infos/Linux-5fdb281ce217f3c395c9f08123080744-2b4bfeacdcbb5d222e1a8671dab48d7a--org.nuxeo.ecm.instance.availability",
         "isTrashed" : false,
         "isProxy" : false,
         "type" : "AdministrativeStatus",
         "changeToken" : "1-0",
         "entity-type" : "document",
         "title" : "Linux-5fdb281ce217f3c395c9f08123080744-2b4bfeacdcbb5d222e1a8671dab48d7a--org.nuxeo.ecm.instance.availability",
         "parentRef" : "41ebabc5-cc66-41d2-9cfb-ad8d3a7287e3",
         "lastModified" : "2020-07-02T14:14:38.276Z",
         "isVersion" : false,
         "repository" : "default"
      }
   ],
   "pageSize" : 1
}
```

This proves that the contribution was successfully loaded.

Shell into the Pod, and perform the following:

```shell
$ nuxeoctl stop
...
Stopping server......Server stopped.
$ nuxeoctl configure
...
Configuration files generated.
$ grep custom.automation.chain.property\
  /opt/nuxeo/server/nxserver/config/configuration.properties
custom.automation.chain.property=123
$ nuxeoctl start
```

Because the Nuxeo CR was defined with a liveness probe that always reports a live status, it is possible to stop and start the Nuxeo server inside the container without Kubernetes recycling the Pod. This is useful in testing/troubleshooting.

The steps above confirmed that the contribution was loaded, and the property defined in the `nuxeo.defaults` key was successfully incorporated into the Nuxeo server configuration.



## Example 2 - PV/PVC

This example has only been tested on Code Ready Containers (CRC.) The steps are similar to those described in the ConfigMap example:

1. Populate a Kubernetes storage resource - in this case a Persistent Volume - with Nuxeo contribution files
2. Create a PVC to reference the Volume
3. Configure Nuxeo to get the contribution files from the PVC instead of a ConfigMap
4. Everything else is the same

CRC comes with a number of preconfigured volumes:

```shell
$ kubectl get pv
NAME       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     ...
pv0001     100Gi      RWO,ROX,RWX    Recycle          Bound      ...
pv0002     100Gi      RWO,ROX,RWX    Recycle          Available  ...
pv0003     100Gi      RWO,ROX,RWX    Recycle          Available  ...
```

For our example, we will use `pv0002` to hold the contribution data. This example is slightly contrived because of the steps required to get the contribution data onto the PV on CRC. A more likely scenario would be a storage type that has a CLI or other interface for moving content onto the Volume. But - this illustrates the configurer's workflow closely enough.

For this example, we will patch the reclaim policy on `pv0002` so that in the process of testing if PVCs are created/deleted the underlying data will persist. However, if you *do* delete the PVC you'll need to manually clear the `PV.Spec.ClaimRef` to re-bind a PVC again. The alternative is leave the policy as *recycle* and then re-copy the contribution files if you delete the PVC. Either way works.

```shell
$ kubectl patch pv pv0002 -p '{"spec":{"persistentVolumeReclaimPolicy":"Retain"}}'
persistentvolume/pv0002 patched
$ kubectl get pv
NAME       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS     ...
...
pv0002     100Gi      RWO,ROX,RWX    Retain           Available  ...
...
```

Here are the detailed steps:

### Populate the PV with contribution files

First, stage the contribution files on the local filesystem. This example will create exactly the same contribution as the ConfigMap example above:

```shell
$ mkdir -p ./volmnt/custom-automation-chain/nxserver/config
$ cd ./volmnt
$ cat <<EOF > ./custom-automation-chain/nuxeo.defaults
custom-automation-chain.target=.
custom.automation.chain.property=123
EOF
$ cat <<EOF > ./custom-automation-chain/nxserver/config/custom-automation-chain-config.xml
<?xml version="1.0"?>
<component name="custom.automation.chain">
  <extension target="org.nuxeo.ecm.core.operation.OperationServiceComponent"
             point="chains">
    <chain id="Document.Query.Extension">
      <operation id="Document.Query">
        <param type="string" name="query">SELECT * FROM Document</param>
        <param type="integer" name="pageSize">1</param>
      </operation>
    </chain>
  </extension>
</component>
EOF
$ tree
.
└── custom-automation-chain
    ├── nuxeo.defaults
    └── nxserver
        └── config
            └── custom-automation-chain-config.xml

3 directories, 2 files
```

The local filesystem is populated. Now, by way of preparation, you can SSH into the CRC VM and explore. This method uses the `-i` option of SSH using an identity file generated by CRC when you spun up the CRC VM:

```shell
$ ssh -i ~/.crc/machines/crc/id_rsa core@$(crc ip)
Red Hat Enterprise Linux CoreOS 43.81.202003310153.0
  Part of OpenShift 4.3, RHCOS is a Kubernetes native operating system
  managed by the Machine Config Operator (`clusteroperator/machine-config`).
...
[core@crc-bq5fv-master-0 ~]$ sudo ls -l /var/mnt/pv-data/
total 0
drwxrwx---. 3 root root 20 Jun  3 14:13 pv0001
drwxrwx---. 2 root root 36 Jul  1 13:49 pv0002
...
```

You can see that the PVs provisioned by CRC map to subdirectories under `/var/mnt/pv-data`. We will use this knowledge to copy the contribution files into the PV. From another local shell, in the same directory where you created the files above, copy the files into the CRC VM.

This command simultaneously creates a tar on the local machine, copies it, and un-tars it in the CRC VM:

```shell
$ tar -c . | ssh -i ~/.crc/machines/crc/id_rsa core@$(crc ip) "sudo tar -x --no-same-owner -C /var/mnt/pv-data/pv0002"
```

From inside the CRC VM, verify:

```shell
$ sudo ls -lR /var/mnt/pv-data/pv0002
/var/mnt/pv-data/pv0002:
total 0
drwxr-xr-x. 3 root root 44 Jul  2 15:21 custom-automation-chain

/var/mnt/pv-data/pv0002/custom-automation-chain:
total 4
-rw-r--r--. 1 root root 70 Jul  2 15:21 nuxeo.defaults
drwxr-xr-x. 3 root root 20 Jul  2 15:07 nxserver

/var/mnt/pv-data/pv0002/custom-automation-chain/nxserver:
total 0
drwxr-xr-x. 2 root root 48 Jul  2 15:09 config

/var/mnt/pv-data/pv0002/custom-automation-chain/nxserver/config:
total 4
-rw-r--r--. 1 root root 440 Jul  2 15:09 custom-automation-chain-config.xml
```

The PV is now populated. You can terminate the SSH session to exit the CRC VM.

### Create a PVC

Create a PVC to gain access to the PV. Note the explicit `volumeName` specification:

```shell
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: pv0002-pvc
spec:
  accessModes:
    - ReadOnlyMany
  resources:
    requests:
      storage: 10M
  volumeMode: Filesystem
  volumeName: pv0002
EOF
persistentvolumeclaim/pv0002-pvc created

$ kubectl get pvc
NAME         STATUS   VOLUME   CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pv0002-pvc   Bound    pv0002   100Gi      RWO,ROX,RWX                   2s

$ kubectl get pv
NAME       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM               ...
...
pv0002     100Gi      RWO,ROX,RWX    Retain           Bound       deleteme/pv0002-pvc ...
```

### Create a Nuxeo CR

This CR produces the same outcome as the ConfigMap CR, but it references the contribution from a PVC, rather than the ConfigMap:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: nuxeo.com/v1alpha1
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
    livenessProbe:
      exec:
        command:
          - "true"
    readinessProbe:
      exec:
        command:
          - "true"
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
    contribs:
    - volumeSource:
        persistentVolumeClaim:
          claimName: pv0002-pvc
      templates:
      - custom-automation-chain
EOF
```

One important difference for this kind of Volume Source is - the Operator mounts the volume directly into `/etc/nuxeo/nuxeo-operator-config` and so only one non-ConfigMap or non-Secret `contribs` entry is supported at this time. So the intent is - when mounting from a persistent  store, all your contributions would be on that store - each in its own subdirectory

For example - if the underlying Volume had multiple contribution peer directories like:

```shell
.
var
└── mnt
    └── pv-data
        └── pv0002
            └── custom-automation-chain
            └── some-other-contribution
            └── third-contribution
            └── this-contribution-ignored
```

Then your Nuxeo CR could look like this:

```shell
apiVersion: nuxeo.com/v1alpha1
kind: Nuxeo
...
  nodeSets:
  - name: cluster
    ...
    contribs:
    - volumeSource:
        persistentVolumeClaim:
          name: pv0002-pvc
      templates:
      - custom-automation-chain
      - some-other-contribution
      - third-contribution
EOF
```

The three named templates would be contributed to the  Nuxeo server, but the fourth one - `this-contribution-ignored` - would **not** be contributed because it is not listed in the `templates` list in the Nuxeo CR. You can see that the template name matches a subdirectory name on the Volume.

It should be noted that **all** of the subdirectories in the Volume are mounted into the container, but the Operator only contributes the items listed in the *templates* list so the other sub-directories are invisible to Nuxeo.

### Verify

You can now perform the same verification steps as in Example One.



## TODO-ME

MicroK8s docs











