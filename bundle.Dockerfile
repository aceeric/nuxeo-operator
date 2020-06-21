FROM scratch

LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=nuxeo-operator
LABEL operators.operatorframework.io.bundle.channels.v1=alpha
LABEL operators.operatorframework.io.bundle.channel.default.v1=alpha

ARG TARGET_CLUSTER=crc
COPY deploy/olm-catalog/nuxeo-operator/manifests/*crd* /manifests/
# todo-me replace with copy/sed
COPY deploy/olm-catalog/nuxeo-operator/manifests/nuxeo-operator.clusterserviceversion.${TARGET_CLUSTER}.yaml /manifests/nuxeo-operator.clusterserviceversion.yaml
COPY deploy/olm-catalog/nuxeo-operator/metadata /metadata/
