FROM fedora:latest

RUN (set -e; \
    dnf install -y make golang dnf-plugins-core git;\
    dnf config-manager addrepo --from-repofile=https://rpm.releases.hashicorp.com/fedora/hashicorp.repo;\
    dnf install -y terraform;\
    git config --global --add safe.directory /home;\
)


WORKDIR /home/
COPY go.mod /home/go.mod
RUN (set -e;\
    go mod tidy;\
)

USER root