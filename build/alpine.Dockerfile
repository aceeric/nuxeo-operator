FROM alpine:3.12.0

ENV OPERATOR=/usr/local/bin/nuxeo-operator \
    USER_UID=1001 \
    USER_NAME=nuxeo-operator

# install operator binary
COPY _output/bin/nuxeo-operator ${OPERATOR}

COPY bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
