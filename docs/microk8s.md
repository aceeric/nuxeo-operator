## Testing the Operator with MicroK8s

In this version of the operator, Kubernetes testing is accomplished using MicroK8s (MK8s) - https://microk8s.io/. The testing was performed on Ubuntu 18.04. Here are the install steps that I followed:

If you need to install *snap*, there are instructions here: https://snapcraft.io/. If you have an existing MK8s snap installed that is not 1.17/stable, remove it. 

```shell
$ sudo snap remove microk8s
```

Install with the 1.17 stable channel:

```shell
$ sudo snap install microk8s --classic --channel=1.17/stable
```

After installation, if you run `sudo microk8s inspect` it will produce:

```
WARNING:  IPtables FORWARD policy is DROP. Consider enabling traffic forwarding with: sudo iptables -P FORWARD ACCEPT 
The change can be made persistent with: sudo apt-get install iptables-persistent
WARNING:  Firewall is enabled. Consider allowing pod traffic with: sudo ufw allow in on cni0 && sudo ufw allow out on cni0
 -- you can back these out with gufw
WARNING:  Docker is installed. 
Add the following lines to /etc/docker/daemon.json: 
{
    "insecure-registries" : ["localhost:32000"] 
}
```

I followed the recommendations. First: `sudo iptables -P FORWARD ACCEPT`. I did not make it permanent because I am reluctant to make permanent changes to networking unless absolutely necessary, so I just remember to set this as needed.

Then: `sudo ufw allow in on cni0 && sudo ufw allow out on cni0`. This can be viewed and removed with `gufw` if/when MK8s is removed. And finally, I added the insecure registry reference into `/etc/docker/daemon.json`  and re-started Docker.

Next, some MK8s add-ons are required:

```shell
$ sudo microk8s enable ingress registry dns dashboard rbac
```

The *registry* add-on enables the internal registry in the Kubernetes cluster for equivalence with the OpenShift testing steps.  The *rbac* add-on is important: without it, the Operator Lifecycle Manager (OLM) cannot be installed into the cluster.

I followed the recommendation and added myself to the `microk8s` group that the snap install creates: `sudo usermod -a -G microk8s $USER`. (For me, this required a restart to take effect.)

With all this configuration complete, I added an alias: `alias mkubectl='microk8s kubectl'`. So in the README, ***mkubectl* is used as the MK8s kubectl**. I did this to keep the existing kubectl separate from the MK8s kubectl.

A quick sanity check:

```shell
$ mkubectl get pod --all-namespaces
NAMESPACE            NAME                                              READY   STATUS  ...
container-registry   registry-d7d7c8bc9-tzjg7                          1/1     Running ...
ingress              nginx-ingress-microk8s-controller-xjn6q           1/1     Running ...
kube-system          coredns-9b8997588-xgsmk                           1/1     Running ...
kube-system          dashboard-metrics-scraper-687667bb6c-25zxg        1/1     Running ...
kube-system          heapster-v1.5.2-5c58f64f8b-74j5c                  4/4     Running ...
kube-system          hostpath-provisioner-7b9cb5cdb4-dsbrg             1/1     Running ...
kube-system          kubernetes-dashboard-5c848cc544-kgff9             1/1     Running ...
kube-system          monitoring-influxdb-grafana-v4-6d599df6bf-4dpwn   2/2     Running ...
```

The main README shows how to configure a Nuxeo instance with a TLS sidecar. This requires TLS Passthrough on the MK8s Nginx ingress controller. Make that configuration change as follows:

```shell
$ mkubectl patch daemonset nginx-ingress-microk8s-controller -n ingress\
  --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--enable-ssl-passthrough"}]'
# verify
$ mkubectl get daemonset -n ingress -o yaml | grep args: -C10
      ...
      spec:
        containers:
        - args:
          - /nginx-ingress-controller
          - --configmap=$(POD_NAMESPACE)/nginx-load-balancer-microk8s-conf
          - --publish-status-address=127.0.0.1
          - --enable-ssl-passthrough
          env:
          ...
```

The test steps in the main README for this project include OLM testing. To install OLM, follow the steps documented here: https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md. My testing was performed with version **0.15.0**. Simplified install steps are:

```shell
$ release=0.15.0
$ url=https://github.com/operator-framework/operator-lifecycle-manager/releases/download/${release}
$ namespace=olm
$ mkubectl apply -f ${url}/crds.yaml
$ mkubectl apply -f ${url}/olm.yaml
```

If the OLM install succeeded, then:

```shell
$ mkubectl get pod -n olm
NAME                               READY   STATUS    ...
catalog-operator-d55f45f8c-d4wpk   1/1     Running   ...
olm-operator-86694fc7f4-9qqng      1/1     Running   ...
operatorhubio-catalog-xl9hn        1/1     Running   ...
packageserver-58c9b8f49d-h4ptk     1/1     Running   ...
packageserver-58c9b8f49d-xt5bl     1/1     Running   ...

$ mkubectl get packagemanifest
NAME                                    CATALOG               AGE
kubemq-operator                         Community Operators   2m29s
banzaicloud-kafka-operator              Community Operators   2m29s
submariner                              Community Operators   2m29s
mattermost-operator                     Community Operators   2m29s
ripsaw                                  Community Operators   2m29s
runtime-component-operator              Community Operators   2m29s
...
```

In order to run the end-to-end tests in MK8s you need to apply a cluster role binding as shown below to enable the operator SDK to initialize the tests. This may have something to do with the fact that the `rbac` add-on is enabled. As an alternative, it would also be possible to not have the operator-sdk create the CRDs, but rather to generate them in the Make file. I may consider that in the future. But for now:

```shell
$ cat <<EOF | mkubectl apply -f -
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: microk8s-e2e
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- apiGroup: rbac.authorization.k8s.io
  kind: User
  name: system:node:$HOSTNAME
EOF
```

Finally, for consistency with the CRC testing, the tests use the same Nuxeo host name in both the MK8s and CRC tests: `nuxeo-server.apps-crc.testing`, even though MK8s runs as a set of processes directly on the local host and CRC runs in a VM. So the Nuxeo server host name must resolve to the MK8s cluster running on your local host. My simple approach is to configure /etc/hosts as shown (this configuration reflects testing on CRC and MK8s alternately):

```shell
...
# CRC
#192.168.130.11  nuxeo-server.apps-crc.testing
# MK8s
127.0.0.1       nuxeo-server.apps-crc.testing
...
```

To remove MK8s and related configurations from your desktop:

1. Uninstall the snap
2. Remove the `microk8s` group
3. Remove the `mkubectl` alias (if you made it persistent)
4. Remove the firewall rules using `gufw`
5. Remove the insecure registry reference from `/etc/docker/daemon.json`
6. Comment out / remove the /etc/hosts entry for the Nuxeo test server host name



------

### Troubleshooting the registry add-on

Since I had to do some registry troubleshooting, I'm adding this documentation for reference. The registry data store can be located this way:

```shell
# find the registry pod
$ mkubectl get pod --all-namespaces
NAMESPACE            NAME                     ...
container-registry   registry-d7d7c8bc9-p4wll ...
...

# find its PVC
$ mkubectl get pod -ncontainer-registry -oyaml | grep volumes: -A3
    volumes:
    - name: registry-data
      persistentVolumeClaim:
        claimName: registry-claim

# find some info about the PVC provisioner
$ mkubectl get pvc --all-namespaces -owide
NAMESPACE          NAME           ... STORAGECLASS        AGE   VOLUMEMODE
container-registry registry-claim ... microk8s-hostpath   12m   Filesystem

$ mkubectl get sc microk8s-hostpath -o yaml | grep provisioner
{"apiVersion":"storage.k8s.io/v1","kind":"StorageClass","metadata":{"annotations":{"storageclass.kubernetes.io/is-default-class":"true"},"name":"microk8s-hostpath"},"provisioner":"microk8s.io/hostpath"}
provisioner: microk8s.io/hostpath

# Take a look at: https://github.com/ubuntu/microk8s/blob/581323a7aa950f10888d4dc8992761d88f09d6a2/microk8s-resources/actions/storage.yaml

# Find the location of the storage on your desktop
$ mkubectl get deployment hostpath-provisioner -nkube-system -oyaml | grep volumes: -A3
      volumes:
      - hostPath:
          path: /var/snap/microk8s/common/default-storage
          type: ""

$ sudo ls -la /var/snap/microk8s/common/default-storage
total 12
drwxr-xr-x 3 root root 4096 Jun 20 14:22 .
drwxr-xr-x 5 root root 4096 Jun 20 14:22 ..
drwxrwxrwx 2 root root 4096 Jun 20 14:22 container-registry-registry-claim-pvc-f2510ecf-ed33-4e3d-be81-22c77d456f61

# as you add images into the registry add-on they will appear in this directory
sudo ls -la /var/snap/microk8s/common/default-storage/container-registry-registry-claim-pvc-f2510ecf-ed33-4e3d-be81-22c77d456f61/docker/registry/v2/

```



