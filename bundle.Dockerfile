FROM scratch

LABEL operators.operatorframework.io.bundle.mediatype.v1=registry+v1
LABEL operators.operatorframework.io.bundle.manifests.v1=manifests/
LABEL operators.operatorframework.io.bundle.metadata.v1=metadata/
LABEL operators.operatorframework.io.bundle.package.v1=nuxeo-operator
LABEL operators.operatorframework.io.bundle.channels.v1=alpha
LABEL operators.operatorframework.io.bundle.channel.default.v1=alpha

COPY deploy/olm-catalog/nuxeo-operator/manifests /manifests/

# the CSV has a replacement token for the operator image. Replace it with a build arg
ARG OPERATOR_IMAGE
RUN sed -i "s|OPERATOR_IMAGE|${OPERATOR_IMAGE}|g" /manifests/nuxeo-operator.clusterserviceversion.yaml
COPY deploy/olm-catalog/nuxeo-operator/metadata /metadata/
