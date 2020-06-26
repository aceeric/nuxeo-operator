# Testing Tomcat as the Reverse Proxy

This README documents the steps to configure Tomcat as the reverse proxy for Nuxeo.

First, generate a Java Keystore containing a self-signed certificate. This will be used by Tomcat to terminate TLS:

```shell
$ keytool -genkey\
 -keyalg RSA\
 -alias tomcat\
 -keystore keystore.jks\
 -storepass tomcatkeystorepassword\
 -dname "C=US, ST=Maryland, L=Somewhere, O=IT, CN=nuxeo-server.apps-crc.testing"
```

Next, generate a secret from the JKS:

```shell
cat <<EOF | oc apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: tomcat-secret
type: Opaque
stringData:
  keystorePass: tomcatkeystorepassword
data:
  keystore.jks: |
$(base64 keystore.jks | sed  's/^/    /')
EOF
```

Verify the secret is correct:

```shell
$ oc extract secret/tomcat-secret --keys=keystorePass --to=-
# keystorePass
tomcatkeystorepassword
$ mkdir deleteme
$ oc extract secret/tomcat-secret --keys=keystore.jks --to=deleteme
$ diff deleteme/keystore.jks ./keystore.jks 
# no output means they are the same
$ rm -rf deleteme
```

With the Nuxeo Operator running (either in-cluster or on the desktop as described in the main README) create a Nuxeo CR that configures Tomcat as the reverse proxy, by virtue of `spec.revProxy.tomcat`:

```shell
cat <<'EOF' | oc apply -f -
apiVersion: nuxeo.com/v1alpha1
kind: Nuxeo
metadata:
  name: my-nuxeo
spec:
  nuxeoImage: image-registry.openshift-image-registry.svc.cluster.local:5000/images/nuxeo:10.10
  access:
    hostname: nuxeo-server.apps-crc.testing
  revProxy:
    tomcat:
      secret: tomcat-secret
  nodeSets:
  - name: cluster
    nuxeoConfig:
      nuxeoPackages:
      - nuxeo-web-ui
    replicas: 1
    interactive: true
    # TODO REMOVE THESE
    livenessProbe:
      exec:
        command:
          - "true"
    readinessProbe:
      exec:
        command:
          - "true"
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

Check the version of Tomcat Nuxeo is

 running:

```
$ oc rsh pod/my-nuxeo-cluster-86b95959cc-pk8nk
$ java -cp /opt/nuxeo/server/lib/catalina.jar org.apache.catalina.util.ServerInfo
Server version: Apache Tomcat/9.0.14
Server built:   Dec 6 2018 21:13:53 UTC
Server number:  9.0.14.0
OS Name:        Linux
OS Version:     4.18.0-147.5.1.el8_1.x86_64
Architecture:   amd64
JVM Version:    1.8.0_252-b09
JVM Vendor:     Oracle Corporation
$ exit
```

