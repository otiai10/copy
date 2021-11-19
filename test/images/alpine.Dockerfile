FROM alpine:3.14

WORKDIR /app

RUN apk update
RUN apk add \
  g++ \
  git \
  go

RUN adduser -g "" -D someuser

# root-owned directory setup for test case 14
RUN mkdir -p test/owned-by-root \
  && chown :someuser test/owned-by-root \
  && chmod 775 test/owned-by-root

USER someuser

COPY --chown=someuser:someuser . .

CMD ["go", "test", "-v", "-cover", "./..."]
