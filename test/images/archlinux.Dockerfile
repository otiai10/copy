FROM archlinux:base-20201220.0.11678

RUN pacman -Sy -q --noconfirm \
  glibc \
  git \
  gcc \
  go

ENV GOPATH=${HOME}/go
ENV GO111MODULE=on
ADD . ${GOPATH}/src/github.com/otiai10/copy
WORKDIR ${GOPATH}/src/github.com/otiai10/copy

CMD ["go", "test", "-v", "-cover", "./..."]
