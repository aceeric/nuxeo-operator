# Connecting Nuxeo to Backing Services

Connecting Nuxeo to a backing service in the cluster - such as Elastic Search - involves three steps for the Nuxeo Operator:



![](../resources/images/backsvc.jpg)



- [x] First, the Operator collects configuration settings from backing service cluster resources
- [x] Second, it projects those resources into the Nuxeo Pod
- [x] Third, it configures nuxeo.conf entries so Nuxeo can connect to the service.

The `spec` section of the Nuxeo CR contains a `backingServices` stanza that supports this. Each item in the backing services list specifies:

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
    nuxeoConf: |
      setting.one=...
      ...
```

Key points:

1. Each backing service has a name that you assign
2. Each backing service specifies one or more in-cluster resources, identified by GVK + Name. For example: `""`, `v1`, `secret`, `elastic-es-http-certs-public`
3. Each resource specifies one or more *projections* that make individual resource configuration values available in the Pod as environment variables or filesystem objects
4. The `nuxeoConf` stanza allows adding nuxeo.conf settings to enable Nuxeo to connect to the backing service. These nuxeo.conf entries are a mix of literals, and references to the projections

### Example One

This example will connect Nuxeo to one backing service: Elastic Cloud on Kubernetes (ECK).

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

Next, create a namespace for a Nuxeo and Elastic Search cluster and deploy an Elastic Search cluster into that namespace. Note the version is 6.8.8 below because according to the Nuxeo documentation (https://doc.nuxeo.com/nxdoc/elasticsearch-setup/#the-rest-client-default) 6.8.X is the latest version supported for Nuxeo LTS 2019 (which includes 10.10):

```shell
$ kubectl create namespace backing
$ cat <<EOF | kubectl apply -n backing -f -
apiVersion: elasticsearch.k8s.elastic.co/v1
kind: Elasticsearch
metadata:
  name: elastic
spec:
  version: 6.8.8
  nodeSets:
  - name: default
    count: 1
    config:
      node.master: true
      node.data: true
      node.store.allow_mmap: false
EOF
```

Give the ECK Operator a few moments, then:

```shell
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

Next, get the elastic cluster TLS cert and use it for a one-way TLS connection:

```shell
# create a work dir to hold the cert
$ mkdir ~/backing
$ cd ~/backing
$ kubectl get secret -nbacking "elastic-es-http-certs-public" -o go-template='{{index .data "tls.crt" | base64decode }}' > tls.crt

$ kubectl port-forward service/elastic-es-http -nbacking 9200 &

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

Port-forwarding simulates an in-cluster connection.

To review, it was shown above that in order to connect to ECK from within the cluster using TLS the following is required:

1. A service name - which is *cluster name*-es-http. So in our case: `elastic-es-http`
2. A user name - which is always `elastic`
3. The elastic user password, which is in a secret named *cluster name*-es-elastic-user. In our case: `elastic-es-elastic-user`. The secret key that contains the password is always `elastic`.
4. A CA certificate, which is also in a secret. The secret is named *cluster name*-es-http-certs-public. So: `elastic-es-http-certs-public`. The certificate is stored in the key `tls.crt`.

Next we will configure a Nuxeo CR to bring up a Nuxeo cluster backed by ECK rather than the internal ElasticSearch that Nuxeo starts up with in development mode.

#### Configure the Nuxeo CR

This example assumes that the Nuxeo Operator is running in the *backing* namespace. This section of the document will build up the Nuxeo CR incrementally. The first backing service to configure is the elastic user secret:

```shell
  backingServices:
  - name: elastic
    resources:
    - group: ""
      version: v1
      kind: secret
      name: elastic-es-elastic-user
      projections:
      - from: elastic
        env: ELASTIC_PASSWORD

```

As shown above, the GVK+Name of the `elastic-es-elastic-user` secret is configured with a projection that makes that password available in the Nuxeo Pod as an environment variable `ELASTIC_PASSWORD`. The `from` element specifies the key within the secret to get the value from.

To reiterate - we know the name of the secret because we created the ElasticSearch cluster with the name `elastic` and so all the dependent resources created by the ECK operator derive from that.

To get the tls.crt certificate from the ECK secret that holds it, another resource and projection is configured:

```shell
  backingServices:
  - name: elastic
    resources:
    - group: ""
      version: v1
      kind: secret
      name: elastic-es-http-certs-public
      projections:
      - transform:
          type: TrustStore
          cert: tls.crt
          store: elastic.ca.jks
          password: elastic.truststore.pass
          passEnv: ELASTIC_TS_PASS
```

This tells the Nuxeo Operator to 1) get the `tls.crt` key from the `elastic-es-http-certs-public` secret, 2) transform it into a Java trust store by way of the `TrustStore` transform, 3) mount the store on the filesystem as `elastic.ca.jks`, 4) generate a trust store password and make it available as an environment variable `ELASTIC_TS_PASS`.

The only remaining task is to configure how Nuxeo should consume these projections. For that, the `nuxeoConf` stanza is used. Below is shown the entire Nuxeo CR incorporating the resources and projections, the `nuxeoConf` stanza, and the other CR components to stand up a Nuxeo cluster. Note the ticks around EOF to preserve the dollar signs in the nuxeo.conf settings:

#### Start Nuxeo

Paste this into the console to cause the Nuxeo Operator to bring up a Nuxeo cluster:

```shell
cat <<'EOF' | kubectl apply -n backing -f -
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
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
  backingServices:
  - name: elastic
    resources:
    - group: ""
      version: v1
      kind: secret
      name: elastic-es-http-certs-public
      projections:
      - transform:
          type: TrustStore
          cert: tls.crt
          store: elastic.ca.jks
          password: elastic.truststore.pass
          passEnv: ELASTIC_TS_PASS
    - group: ""
      version: v1
      kind: secret
      name: elastic-es-elastic-user
      projections:
      - key: elastic
        env: ELASTIC_PASSWORD
    nuxeoConf: |
      elasticsearch.client=RestClient
      elasticsearch.restClient.username=elastic
      elasticsearch.restClient.password=${env:ELASTIC_PASSWORD}
      elasticsearch.addressList=https://elastic-es-http:9200
      elasticsearch.restClient.truststore.path=/etc/nuxeo-operator/binding/elastic/elastic.ca.jks
      elasticsearch.restClient.truststore.password=${env:ELASTIC_TS_PASS}
      elasticsearch.restClient.truststore.type=JKS
EOF
```



#### Verify

Confirm the Nuxeo Pod is running:

```shell
$ kubectl get po -nbacking
NAME                                READY     STATUS    RESTARTS   AGE
elastic-es-default-0                1/1       Running   0          82m
my-nuxeo-cluster-78bd6dfcf7-lxnkf   1/1       Running   0          15s
```

Shell into the Nuxeo pod:

```shell
$ kubectl exec -it my-nuxeo-cluster-78bd6dfcf7-lxnkf bash
```

Then from within the pod, query the Elastic Search indexes:

```shell
$ nuxeo@my-nuxeo-cluster-78bd6dfcf7-lxnkf:~$ curl -k -u elastic:$ELASTIC_PASSWORD https://elastic-es-http:9200/_aliases?pretty=true
{
  "nuxeo" : {
    "aliases" : { }
  },
  "nuxeo-audit" : {
    "aliases" : { }
  },
  "nuxeo-uidgen" : {
    "aliases" : { }
  }
}
```

You can see that Nuxeo has automatically created its required ElasticSearch indexes as part of its startup. You can also see that the elastic user password environment variable is correctly defined.

Examine the Nuxeo logs. The logs should look something like:

```shell
$nuxeo@my-nuxeo-cluster-78bd6dfcf7-lxnkf:~$ cat /var/log/nuxeo/server.log
...
======================================================================
= Component Loading Status: Pending: 0 / Missing: 0 / Unstarted: 0 / Total: 499
======================================================================
2020-07-17T20:11:04,113 WARN  [I/O dispatcher 1] [org.elasticsearch.client.RestClient] request [POST https://elastic-es-http:9200/nuxeo/_search?typed_keys=true&ignore_unavailable=false&expand_wildcards=open&allow_no_indices=true&search_type=dfs_query_then_fetch&batched_reduce_size=512] returned 1 warnings: [299 Elasticsearch-6.8.8-2f4c224 "'y' year should be replaced with 'u'. Use 'y' for year-of-era. Prefix your date format with '8' to use the new specifier."]
...
```

The key point is - all components should be started and there should be ElasticSearch chatter in the logs.

#### Review

You can see that the nuxeoConf stanza has some literal settings like `elasticsearch.client=RestClient`, and other settings that reference environment variables and filesystem mounts. Notice that anything projected as a mount is relative to `/etc/nuxeo-operator/binding/elastic`.

All backing service mounts are relative to `etc/nuxeo-operator/binding` and then each specific binding - in this case "elastic" - is mounted separately under that. Then each projected file is mounted within that directory.

All backing service nuxeo.conf settings are combined together, along with any inline nuxeo.conf settings specified in the main CR body, and placed into a ConfigMap managed by the Operator. The ConfigMap is named *the nuxeo cluster name*-*the nodeSet name*-nuxeo-conf. E.g.:

```shell
$ kubectl get cm my-nuxeo-cluster-nuxeo-conf -oyaml
apiVersion: v1
data:
  nuxeo.conf: |
    elasticsearch.client=RestClient
    elasticsearch.restClient.username=elastic
    elasticsearch.restClient.password=${env:ELASTIC_PASSWORD}
    elasticsearch.addressList=https://elastic-es-http:9200
    elasticsearch.restClient.truststore.path=/etc/nuxeo-operator/binding/elastic/elastic.ca.jks
    elasticsearch.restClient.truststore.password=${env:ELASTIC_TS_PASS}
    elasticsearch.restClient.truststore.type=PKCS12
kind: ConfigMap
metadata:
  name: my-nuxeo-cluster-nuxeo-conf
  namespace: backing
  ...
```

Whenever a transform of an upstream resource is required, as was the case for the trust store, the Operator has to create a *secondary secret* to hold the transformed value. The secret is named *nuxeo cluster name*-secondary-*backing name*. E.g.:

```shell
$ kubectl describe secret my-nuxeo-secondary-elastic
Name:         my-nuxeo-secondary-elastic
Namespace:    backing
Labels:       <none>
Annotations:  nuxeo.operator.v1.secret.elastic-es-http-certs-public: 2189411

Type:  Opaque

Data
====
elastic.ca.jks:           1820 bytes
elastic.truststore.pass:  12 bytes
```

You can see that the operator placed the Java trust store, and the Operator-generated trust store password into this Secret. In addition, the secret was annotated with the resource version of the TLS cert secret that the trust store was created from.

#### For the patient reader

The binding approach presented above provides insight into how the Nuxeo Operator binds Nuxeo to a backing service. However, that binding is fairly verbose, especially if you're running Nuxeo with multiple backing services. There is a simpler way.

The simple way is based on understanding that the operator-managed backing services like Strimzi and ECK are machines and - as such - they always do the same thing. The Nuxeo Operator can take advantage of that with something called *pre-configured* bindings.

Pre-configured bindings are bindings to backing services like ECK, Strimzi Kafka, and Crunchy Postgres pre-wired into the Operator. For ECK, for example, the following Nuxeo CR will perform exactly the same binding as the more verbose example above:

```shell
cat <<'EOF' | kubectl apply -n backing -f -
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
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
  backingServices:
  - preConfigured:
      type: ECK
      resource: elastic
EOF
```

You can see that the `backingServices` stanza simply provides a `preConfigured` entry, specifying type `ECK`, which is one of the supported pre-configured backing services. The only additional piece of information needed is the `resource` key, which specifies that in the namespace that the Nuxeo cluster is running in, there is instance of a `elasticsearch.k8s.elastic.co` resource named `elastic`.

That's all that's required by the Operator to bind Nuxeo to ECK. Feel free to try it.

The current version of the Nuxeo Operator supports ECK, Strimzi, and Crunchy Postgres.

