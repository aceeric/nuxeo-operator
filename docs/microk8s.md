## Installing MicroK8s

In this version of the operator, Kubernetes testing is accomplished using MicroK8s - https://microk8s.io/ - on Ubuntu 18.04. Here are the install steps that were followed:

If you need to install snap, there are instructions elsewhere for that. If you have an existing MicroK8s snap installed that is not 1.17/stable, remove it. 

```shell
$ sudo snap remove microk8s
```

Install with the 1.17 stable channel:

```shell
$ sudo snap install microk8s --classic --channel=1.17/stable
```

The reason to use 1.17 is documented here: https://kubernetes.io/docs/setup/release/notes/#other-api-changes in the bullet that talks about *CustomResourceDefinition schemas that use `x-kubernetes-list-map-keys`*. By default, MicroK8s installs with Kubernetes 1.18, and that version handles validation differently. This affects the Nuxeo CR, because it as of the current release it includes a `PodTemplate` spec. After installation, if you run `sudo microk8s inspect` it will produce:

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

I followed the recommendations. First: `sudo iptables -P FORWARD ACCEPT`. I did **not** make it permanent because I am reluctant to make permanent changes to networking unless absolutely necessary, so I just remember to set this on each reboot.

Then: `sudo ufw allow in on cni0 && sudo ufw allow out on cni0`. This can be viewed and removed with `gufw` if/when MicroK8s is removed. And finally, I added the insecure registry reference and re-started Docker. Then:

```
$ sudo microk8s enable ingress registry dns dashboard
```

This enables the internal registry in the Kubernetes cluster for equivalence with the OpenShift testing steps.  Finally, I followed their recommendation and added myself to the `microk8s` group that the snap install creates: `sudo usermod -a -G microk8s $USER`. On Ubuntu, this requires a restart...

With all this configuration complete, I added an alias: `alias mkubectl='microk8s kubectl'`. I did this to keep the existing kubectl separate from the MicroK8s kubectl. Then:

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

To remove MicroK8s:

1. Uninstall the snap
2. Remove the `microk8s` group
3. Remove the `mkubectl` alias (if you bothered making it persistent)
4. Remove the firewall rules using `gufw`
5. Remove the insecure registry reference from `/etc/docker/daemon.json`

