# Testing JVM PKI configuration

This README documents the steps to configure Nuxeo with a JVM-wide Key Store and Trust Store. This feature supports a use case where you run Nuxeo with a custom marketplace package which requires the JVM to be configured with an externally provisioned key store and trust store. Perhaps your custom package uses a legacy library that can only use the JVM SSL configuration.

Create a workspace on your local file system:
```shell
$ mkdir jvmpki
$ cd jvmpki
```

Create a key store:
```shell
$ keytool -genkey\
 -keyalg RSA\
 -alias nuxeo\
 -keystore keystore.p12\
 -storetype pkcs12\
 -storepass nuxeokeystorepassword\
 -keypass nuxeokeystorepassword\
 -dname "C=US, ST=Maryland, L=Somewhere, O=IT, CN=nuxeo-server.apps-crc.testing"
```

Export the certificate from the key store:
```shell
$ keytool -export\
 -alias nuxeo\
 -file certificate.cer\
 -storepass nuxeokeystorepassword\
 -keystore keystore.jks
```

Create the trust store and import the certificate into it:
```shell
$ keytool -import\
 -v -trustcacerts\
 -alias nuxeo\
 -file certificate.cer\
 -storetype pkcs12\
 -storepass nuxeotruststorepassword\
 -noprompt\
 -keystore truststore.p12
```

Generate a secret incorporating the two stores and passwords and store types:
```shell
$ cat <<EOF | oc apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: jvm-pki-secret
type: Opaque
stringData:
  keyStoreType: PKCS12
  keyStorePassword: nuxeokeystorepassword
  trustStoreType: PKCS12
  trustStorePassword: nuxeotruststorepassword
data:
  keyStore: |
$(base64 keystore.p12 | sed  's/^/    /')
  trustStore: |
$(base64 truststore.p12 | sed  's/^/    /')
EOF
```

With the Nuxeo Operator running (either in-cluster or on the desktop as described in the main README) create a Nuxeo CR that configures Nuxeo for JVM PKI using the secret created above:
```shell
$ cat <<EOF | oc apply -f -
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
      nuxeoPackages:
      - nuxeo-web-ui
      # this property configures Nuxeo:
      jvmPKISecret: jvm-pki-secret
EOF
```

Wait for the Nuxeo Pod to be running:

```shell
$ oc get pod
  NAME                                READY   STATUS        RESTARTS   AGE
  my-nuxeo-cluster-66c69959f8-hksb2   1/1     Running       0          50s
```

Shell into the pod
```shell
$ oc rsh my-nuxeo-cluster-66c69959f8-hksb2
```

Inside the pod, examine the configuration
```shell
# examine the key store

$ keytool -list -v -keystore /etc/pki/jvm/keystore.p12 -storepass nuxeokeystorepassword | head
Keystore type: PKCS12
Keystore provider: SUN

Your keystore contains 1 entry

Alias name: nuxeo
Creation date: Jun 28, 2020
Entry type: PrivateKeyEntry
Certificate chain length: 1
Certificate[1]:

# examine the trust store

$ keytool -list -v -keystore /etc/pki/jvm/truststore.p12 -storepass nuxeotruststorepassword | head
Keystore type: PKCS12
Keystore provider: SUN

Your keystore contains 1 entry

Alias name: nuxeo
Creation date: Jun 28, 2020
Entry type: trustedCertEntry

Owner: C=US, ST=Maryland, L=Somewhere, O=IT, CN=nuxeo-server.apps-crc.testing

# examine JAVA_OPTS

$ echo $JAVA_OPTS
-XX:+UnlockExperimentalVMOptions -XX:+UseCGroupMemoryLimitForHeap -XX:MaxRAMFraction=1
 -Djavax.net.ssl.keyStoreType=PKCS12 -Djavax.net.ssl.keyStore=/etc/pki/jvm/keystore.p12
 -Djavax.net.ssl.keyStorePassword=nuxeokeystorepassword -Djavax.net.ssl.trustStoreType=PKCS12 
 -Djavax.net.ssl.trustStore=/etc/pki/jvm/truststore.p12 
 -Djavax.net.ssl.trustStorePassword=nuxeotruststorepassword

# Get the running Nuxeo process

$ ps -eaf
UID          PID    PPID  C STIME TTY          TIME CMD
...
nuxeo        157     143 16 00:36 ?        00:00:30 /usr/local/openjdk-8/jre
...

# examine the args that Nuxeo was started with

$ xargs -0 printf '%s\n' </proc/157/cmdline
...
-Djavax.net.ssl.keyStoreType=PKCS12
-Djavax.net.ssl.keyStore=/etc/pki/jvm/keystore.p12
-Djavax.net.ssl.keyStorePassword=nuxeokeystorepassword
-Djavax.net.ssl.trustStoreType=PKCS12
-Djavax.net.ssl.trustStore=/etc/pki/jvm/truststore.p12
-Djavax.net.ssl.trustStorePassword=nuxeotruststorepassword
...
```
