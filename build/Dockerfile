FROM foo

# operator-sdk expects this file but my project does not use it and it is not a real docker file
#  see https://github.com/operator-framework/operator-sdk/issues/1804#issuecomment-542329479
#  (may be obviated by osdk v1.0.0)

ENV OPERATOR=/usr/local/bin/nuxeo-operator \
    USER_UID=1001 \
    USER_NAME=nuxeo-operator

# install operator binary
COPY _output/bin/nuxeo-operator ${OPERATOR}

COPY bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
