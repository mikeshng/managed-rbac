FROM registry.ci.openshift.org/stolostron/builder:go1.20-linux AS builder

WORKDIR /go/src/github.com/open-cluster-management-io/managed-rbac
COPY . .
RUN make -f Makefile build

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

RUN microdnf update && \
     microdnf clean all

ENV OPERATOR=/usr/local/bin/managed-rbac \
    USER_UID=1001 \
    USER_NAME=managed-rbac

# install operator binary
COPY --from=builder /go/src/github.com/open-cluster-management-io/managed-rbac/bin/managed-rbac /usr/local/bin/managed-rbac

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
