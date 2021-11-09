FROM alpine:3.14

RUN apk update
RUN apk add \
  g++ \
  git \
  go

ENV GOPATH=${HOME}/go
ENV GO111MODULE=on
ADD . ${GOPATH}/src/github.com/otiai10/copy
WORKDIR ${GOPATH}/src/github.com/otiai10/copy

CMD ["go", "test", "-v", "-cover", "./..."]
