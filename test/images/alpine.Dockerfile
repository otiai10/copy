FROM alpine:3.14

WORKDIR /app

RUN apk update
RUN apk add \
  g++ \
  git \
  go

RUN adduser -g "" -D someuser

USER someuser

COPY --chown=someuser:someuser . .

CMD ["go", "test", "-v", "-cover", "./..."]
