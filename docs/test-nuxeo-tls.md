# Testing TLS Termination directly in Nuxeo

This README documents the steps to configure Nuxeo to terminate TLS. These steps are presented for testing only.

First, generate a Java Keystore containing a self-signed certificate:

```shell
$ keytool -genkey\
 -keyalg RSA\
 -alias tomcat\
 -keystore keystore.jks\
 -storepass nuxeokeystorepassword\
 -dname "C=US, ST=Maryland, L=Somewhere, O=IT, CN=nuxeo-server.apps-crc.testing"
```

Note - the alias 'tomcat' is expected by Nuxeo as of 10.10. Next, generate a secret from the JKS. The secret must contain keys 'keystore.jks' and 'keystorePass':

```shell
cat <<EOF | oc apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: tls-secret
type: Opaque
stringData:
  keystorePass: nuxeokeystorepassword
data:
  keystore.jks: |
$(base64 keystore.jks | sed  's/^/    /')
EOF
```

Verify the secret is correct:

```shell
$ oc extract secret/tls-secret --keys=keystorePass --to=-
# keystorePass
nuxeokeystorepassword
$ mkdir deleteme
$ oc extract secret/tls-secret --keys=keystore.jks --to=deleteme
$ diff deleteme/keystore.jks ./keystore.jks 
# no output means they are the same
$ rm -rf deleteme
```

With the Nuxeo Operator running (either in-cluster or on the desktop as described in the main README) create a Nuxeo CR that configures Nuxeo for TLS:

```shell
cat <<'EOF' | oc apply -f -
apiVersion: appzygy.net/v1alpha1
kind: Nuxeo
metadata:
  name: my-nuxeo
spec:
  nuxeoImage: nuxeo:LTS-2019
  access:
    hostname: nuxeo-server.apps-crc.testing
  nodeSets:
  - name: cluster
    replicas: 1
    interactive: true
    nuxeoConfig:
      # this causes the Operator to configure Nuxeo for TLS
      tlsSecret: tls-secret
      nuxeoPackages:
      - nuxeo-web-ui
EOF
```

Verify:

```shell
$ oc get route,pod
NAME                                              HOST/PORT                     ...
route.route.openshift.io/my-nuxeo-cluster-route   nuxeo-server.apps-crc.testing ...

NAME                                    READY   STATUS    RESTARTS   AGE
pod/my-nuxeo-cluster-86b95959cc-pk8nk   1/1     Running   0          3h51m
```

Then, access the server from your browser with the URL: `https://nuxeo-server.apps-crc.testing`

Check the version of Tomcat Nuxeo is running:

```shell
$ oc exec my-nuxeo-cluster-86b95959cc-pk8nk -- \
  java -cp /opt/nuxeo/server/lib/catalina.jar org.apache.catalina.util.ServerInfo
Server version: Apache Tomcat/9.0.14
Server built:   Dec 6 2018 21:13:53 UTC
Server number:  9.0.14.0
OS Name:        Linux
OS Version:     4.18.0-147.5.1.el8_1.x86_64
Architecture:   amd64
JVM Version:    1.8.0_252-b09
JVM Vendor:     Oracle Corporation
```

