FROM fedora:latest

RUN (set -e; \
    dnf install -y make golang dnf-plugins-core;\
    dnf config-manager addrepo --from-repofile=https://rpm.releases.hashicorp.com/fedora/hashicorp.repo;\
    dnf install -y terraform;\
)

WORKDIR /home/

USER root