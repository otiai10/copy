FROM centos:centos8

RUN yum update -y -q \
    && yum install -y --quiet \
      git \
      go

ENV GOPATH=${HOME}/go
ENV GO111MODULE=on
ADD . ${GOPATH}/src/github.com/otiai10/copy
WORKDIR ${GOPATH}/src/github.com/otiai10/copy

CMD ["go", "test", "-v", "-cover", "./..."]
