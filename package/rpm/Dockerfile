FROM docker.io/library/rockylinux:9

# Install rpm requirements
RUN dnf -y install git curl tar rpm-build rpmdevtools

# Install Golang
ARG GOLANG_VERSION=1.21.5
ARG GOLANG_ARCH=amd64
WORKDIR /usr/local
RUN curl -L https://dl.google.com/go/go${GOLANG_VERSION}.linux-${GOLANG_ARCH}.tar.gz | tar -xzf -
RUN ln -s /usr/local/go/bin/go /usr/bin/go
RUN ln -s /usr/local/go/bin/gofmt /usr/bin/gofmt

# Build RPM
WORKDIR /root/rpmbuild
ADD configsync.spec .
RUN mkdir SOURCES
COPY configsync-*.tar.gz SOURCES/
ADD entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

VOLUME /root/rpmbuild/RPMS

ENTRYPOINT /entrypoint.sh
