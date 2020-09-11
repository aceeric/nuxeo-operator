# Nuxeo Operator metrics

The Nuxeo Operator is scaffolded using the Operator SDK v1.0.0. The scaffolding includes default support for Prometheus metrics. To verify the functionality of the metrics, perform the following steps.

These steps assume you have git cloned the Operator repo and are logged into a Kubernetes cluster with cluster admin privileges.

### Install the Nuxeo Operator in test mode

```shell
$ make operator-install 
No resources found
namespace/nuxeo-operator-system created
customresourcedefinition.apiextensions.k8s.io/nuxeos.appzygy.net created
serviceaccount/nuxeo-operator-manager created
role.rbac.authorization.k8s.io/nuxeo-operator-leader-election created
clusterrole.rbac.authorization.k8s.io/nuxeo-operator-manager created
clusterrole.rbac.authorization.k8s.io/nuxeo-operator-proxy created
clusterrole.rbac.authorization.k8s.io/nuxeo-operator-metrics-reader created
rolebinding.rbac.authorization.k8s.io/nuxeo-operator-leader-election created
clusterrolebinding.rbac.authorization.k8s.io/nuxeo-operator-manager created
clusterrolebinding.rbac.authorization.k8s.io/nuxeo-operator-proxy created
service/nuxeo-operator-metrics-service created
deployment.apps/nuxeo-operator-controller-manager created
```
The `operator-install` Make target creates a namespace `nuxeo-operator-system` and installs the Operator and RBACs into that namespace, and the CRD into the cluster.

### Ensure the Operator is running successfully

Verify the Nuxeo Operator is running:

```shell
$ kubectl get pod -n nuxeo-operator-system
NAME                                                 READY     STATUS    RESTARTS   AGE
nuxeo-operator-controller-manager-59df968b57-lpn9j   2/2       Running   0          16s
```

### Port-forward to the Operator pod in one shell

Since there is no Route or Ingress for the Operator Pod, use port-forwarding to access the metrics endpoint (8443) on the Operator Pod:

```shell
$ kubectl port-forward -n nuxeo-operator-system\
  nuxeo-operator-controller-manager-59df968b57-lpn9j 8443
Forwarding from 127.0.0.1:8443 -> 8443
Forwarding from [::1]:8443 -> 8443
```
Use port 8443 because that's the port that the metrics are published to. (The *kube-auth-proxy* sidecar in the Operator Pod exposes the metrics.)

Next, get your token. You need a token because the metrics *kube-auth-proxy* uses RBAC to authenticate and authorize a client's metrics request:
```shell
# OpenShift
$ oc whoami -t
DEd1xOB5lBRLYq4N2gwiBths4mvOpBfQo_20SiLOyuo
# Kuberbetes
TODO GET AN UBER TOKEN IN K8S
```
### Access the metrics via curl in another shell

In another shell, use curl with your token to see the metrics:

```shell
$ curl -k https://localhost:8443/metrics\
  -H "Accept: application/json"\
  -H "Authorization: Bearer DEd1xOB5lBRLYq4N2gwiBths4mvOpBfQo_20SiLOyuo"
# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 7.198e-06
go_gc_duration_seconds{quantile="0.25"} 3.055e-05
go_gc_duration_seconds{quantile="0.5"} 6.883e-05
go_gc_duration_seconds{quantile="0.75"} 0.000689496
go_gc_duration_seconds{quantile="1"} 0.010060858
go_gc_duration_seconds_sum 0.01283921
go_gc_duration_seconds_count 12
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 81
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.13.15"} 1
# HELP go_memstats_alloc_bytes Number of bytes allocated and still in use.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 7.9765688e+07
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated, even if freed.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 2.13977776e+08
# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table.
# TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 1.475731e+06
# HELP go_memstats_frees_total Total number of frees.
# TYPE go_memstats_frees_total counter
go_memstats_frees_total 213336

<remainder deleted for brevity>
```

You can see that the curl request returned a lengthy text output of metric information. If you're familiar with Prometheus metrics, this format will look familiar. If you've never looked at raw Prometheus metrics, they consist of two `# HELP` lines, following by a set of metrics.

### Verify the Metrics in Prometheus

The example below shows the instructions for verifying Prometheus metrics in Code Ready Containers v13. According to [this documentation](https://code-ready.github.io/crc/#administrative-tasks_gsg), Prometheus and the related monitoring, alerting, and telemetry are disabled in CRC by default, and the link has instructions for enabling telemetry.

If you're using another Kubernetes/OpenShift environment then enabling telemetry may be different or it may already have been enabled. With telemetry enabled, you might see something like:

```shell
$ kubectl get po -n openshift-monitoring
NAME                                          READY     STATUS    RESTARTS   AGE
alertmanager-main-0                           5/5       Running   0          60s
alertmanager-main-1                           5/5       Running   0          60s
alertmanager-main-2                           5/5       Running   0          60s
cluster-monitoring-operator-f76d748f6-6cmdr   2/2       Running   0          82s
grafana-f457c8645-g5b2s                       2/2       Running   0          48s
kube-state-metrics-5b557cf9c6-6pnwv           3/3       Running   0          73s
node-exporter-sc86d                           2/2       Running   0          67s
openshift-state-metrics-7db99f498c-bvlb2      3/3       Running   0          72s
prometheus-adapter-5596d657c8-zn5ft           1/1       Running   0          52s
prometheus-adapter-5596d657c8-zrc8k           1/1       Running   0          52s
prometheus-k8s-0                              7/7       Running   1          50s
prometheus-k8s-1                              7/7       Running   1          50s
prometheus-operator-66f6479d8c-vfk2l          2/2       Running   0          70s
telemeter-client-596fbf7b4d-pzzjq             3/3       Running   0          57s
thanos-querier-755f6677b6-6ndtq               4/4       Running   0          53s
thanos-querier-755f6677b6-pk7fq               4/4       Running   0          53s
```

Then there are a couple of ways to access the metrics. Again - this is CRC-specific. With `crc console`, from the `Monitoring > Metrics` link in the console sidebar,  You can click the `Prometheus UI` link which takes you to the Prometheus UI. Or:

```shell
$ kubectl get route -n openshift-monitoring
NAME                HOST/PORT                                             ...
...
prometheus-k8s      prometheus-k8s-openshift-monitoring.apps-crc.testing  ...
...
```

Then, in a browser tab, access `https://prometheus-k8s-openshift-monitoring.apps-crc.testing`. Login with `kube:admin`. Supply credentials from `crc console --credentials`. 

In CRC, Prometheus immediately had access to the Nuxeo Operator metrics. For example, using the following Prometheus query: `pod:container_cpu_usage:sum{namespace="nuxeo-operator-system",pod="nuxeo-operator-controller-manager-59df968b57-lpn9j"}`, the following graph is produced  in Prometheus without requiring any additional configuration:

![](../resources/images/prometheus.png)

The Nuxeo Operator installs with a `ClusterRole` named `nuxeo-operator-metrics-reader` that uses RBAC to enable access to the metrics endpoint in the Nuxeo Operator container. It's possible that you might have to add a `ClusterRoleBinding` to bind the Prometheus service account to this `ClusterRole` as documented [here](https://book.kubebuilder.io/reference/metrics.html). But - this is likely environment-dependent.

Either way - this should demonstrate that the Nuxeo Operator is configured for Prometheus metric exposition out of the box.



