FROM golang:1.27-alpine

WORKDIR /app

RUN adduser -g "" -D someuser

# Root-owned directory for test case 14
RUN mkdir -p test/owned-by-root \
  && chown :someuser test/owned-by-root \
  && chmod 775 test/owned-by-root

USER someuser

COPY --chown=someuser:someuser . .

CMD ["go", "test", "-v", "-cover", "./..."]
