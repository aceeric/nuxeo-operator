# Connecting Nuxeo to Backing Services

Connecting Nuxeo to a backing service in the cluster such as Elastic Search etc. involves three steps for the Nuxeo Operator:

![](../resources/images/backsvc.jpg)

- [x] First, the Operator collects configuration settings from backing service cluster resources
- [x] Second, it projects those resources into the Nuxeo Pod
- [x] Third, it configures nuxeo.conf entries so Nuxeo can connect to the service.

The `spec` section of the Nuxeo CR contains a `backingServices` entry that supports this. Each item in the backing services list specifies:

1. Cluster resources to collect configuration values from

2. How to project those configuration values into the Nuxeo Pod

3. How to configure Nuxeo to use the projected values

The high level structure of the `backingsServices` stanza is shown here:

```shell
apiVersion: nuxeo.com/v1alpha1
kind: Nuxeo
metadata:
  name: my-nuxeo
spec:
  backingServices:
  - name:
    resources:
    - group:
      version:
      kind:
      name:
      projections:
    ...
    nuxeo.conf: |
      setting.one=...
      ...
```

Key points:

1. Each backing service has a name that you assign
2. Each backing service specifies one or more resources, identified by GVK + Name. For example: `""`, `v1`, `secret`, `elastic-es-http-certs-public`
3. Each resource specifies one or more *projections* that make individual resource configuration values available in the Pod as environment variables or filesystem objects
4. The nuxeo.conf section has entries to add to nuxeo.conf to enable Nuxeo to connect to the backing service. These nuxeo.conf entries are a mix of literals, and references to the projections

### Example One

This example will install one backing service: Elastic Cloud on Kubernetes (ECK). The environment used for this examples is MicroK8s. The steps are:

1. **Install** - Install ECK into the Kubernetes cluster
2. **Connect** - Manually connect to Elastic Search from outside of the cluster to verify the installation and also to understand the elements that are needed for a client connection
3. **Configure Nuxeo** - Configure a Nuxeo CR to utilize ECK using the knowledge gained in step two
4. **Start Nuxeo** - Start the Nuxeo cluster
5. **Verify** - Confirm that Nuxeo is connected to ECK from within the cluster
6. **Summarize** - Review the example and highlight specific aspects of the Nuxeo CR backing service stanza

#### Install

Begin by creating a new namespace `elastic-system`, and installing the ECK operator into that namespace. These steps are based on the ECK quick start at https://www.elastic.co/guide/en/cloud-on-k8s/current/k8s-quickstart.html:

```shell
$ kubectl create namespace elastic-system
$ kubectl apply -n elastic-system -f https://download.elastic.co/downloads/eck/1.1.2/all-in-one.yaml
```

Next, create a namespace for an Elastic Search cluster and deploy an Elastic Search cluster into that namespace:

```shell
$ kubectl create namespace backing
$ cat <<EOF | kubectl apply -n backing -f -
apiVersion: elasticsearch.k8s.elastic.co/v1
kind: Elasticsearch
metadata:
  name: elastic
spec:
  version: 7.8.0
  nodeSets:
  - name: default
    count: 1
    config:
      node.master: true
      node.data: true
      node.store.allow_mmap: false
EOF

$ kubectl get elasticsearch -nbacking
NAME      HEALTH   NODES   VERSION   PHASE   AGE
elastic   green    1       7.8.0     Ready   81s
```

#### Connect

This example will follow the steps outlined on the ECK documentation: https://www.elastic.co/guide/en/cloud-on-k8s/current/k8s-accessing-elastic-services.html and https://www.elastic.co/guide/en/cloud-on-k8s/current/k8s-request-elasticsearch-endpoint.html.

Initially, get the *elastic* user password, port-forward to the elastic service, and use `curl` in insecure mode to access the endpoint:

```sh
$ kubectl get service elastic-es-http -nbacking
NAME              TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
elastic-es-http   ClusterIP   10.152.183.7   <none>        9200/TCP   19h
$ PASSWORD=$(kubectl get secret elastic-es-elastic-user -nbacking\
  -o go-template='{{.data.elastic | base64decode}}')
$ kubectl port-forward service/elastic-es-http -nbacking 9200 &
$ curl -u elastic:$PASSWORD -k https://localhost:9200
{
  "name" : "elastic-es-default-0",
  "cluster_name" : "elastic",
  "cluster_uuid" : "gctomTeHRMyimlOF8u24mQ",
  "version" : {
    "number" : "7.8.0",
    "build_flavor" : "default",
    "build_type" : "docker",
    "build_hash" : "757314695644ea9a1dc2fecd26d1a43856725e65",
    "build_date" : "2020-06-14T19:35:50.234439Z",
    "build_snapshot" : false,
    "lucene_version" : "8.5.1",
    "minimum_wire_compatibility_version" : "6.8.0",
    "minimum_index_compatibility_version" : "6.0.0-beta1"
  },
  "tagline" : "You Know, for Search"
}
```

Next, get the elastic cluster CA cert and use it for a one-way TLS connection:

```shell
# create a work dir to hold the cert
$ mkdir ~/backing
$ cd ~/backing
$ kubectl get secret -nbacking "elastic-es-http-certs-public" -o go-template='{{index .data "tls.crt" | base64decode }}' > tls.crt

# this example uses a bit of a gimmick. I added an entry to /etc/hosts with the elastic
# service "host" name that routes to the localhost. Again, port-forward to the service.
# The result is that the curl request pases the host name that matches the cert, and so
# curl won't complain about the using localhost to access elastic-es-http hostname
$ cat /etc/hosts
...
127.0.0.1       elastic-es-http
...

$ mkubectl port-forward service/elastic-es-http -nbacking 9200

# this time specify the CA obtained from Elastic Search
$ curl --cacert tls.crt -u elastic:$PASSWORD https://elastic-es-http:9200/
{
  "name" : "elastic-es-default-0",
  "cluster_name" : "elastic",
  "cluster_uuid" : "gctomTeHRMyimlOF8u24mQ",
  "version" : {
    "number" : "7.8.0",
    "build_flavor" : "default",
    "build_type" : "docker",
    "build_hash" : "757314695644ea9a1dc2fecd26d1a43856725e65",
    "build_date" : "2020-06-14T19:35:50.234439Z",
    "build_snapshot" : false,
    "lucene_version" : "8.5.1",
    "minimum_wire_compatibility_version" : "6.8.0",
    "minimum_index_compatibility_version" : "6.0.0-beta1"
  },
  "tagline" : "You Know, for Search"
}
```

Port-forwarding simulates an in-cluster connection. To review, it has been shown above that in order to connect to ECK from within the cluster using TLS the following is required:

1. A service name - which is *cluster name*-es-http. So in our case: `elastic-es-http`
2. A user name - which is always `elastic`
3. A password, which is in a secret named *cluster name*-es-elastic-user. In our case: `elastic-es-elastic-user`. The secret key that contains the password is `elastic`.
4. A CA certificate, which is also in a secret. In this case, the secret is *cluster name*-es-http-certs-public. So: `elastic-es-http-certs-public`. The certificate is stored in the key `tls.crt`.

Next we will configure a Nuxeo CR to bring up a Nuxeo cluster with ECK rather than the internal ElasticSearch that it comes up with in development mode.

#### Configure Nuxeo

This section of the document will build up the Nuxeo CR incrementally.





Start Nuxeo



Verify



Summarize







