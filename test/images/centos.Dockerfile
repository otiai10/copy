FROM centos:centos8

WORKDIR /app

RUN yum update -y -q \
    && yum install -y --quiet \
      git \
      go

RUN useradd -m someuser

# root-owned directory setup for test case 14
RUN mkdir -p test/owned-by-root \
  && chown :someuser test/owned-by-root \
  && chmod 775 test/owned-by-root

USER someuser

COPY --chown=someuser:someuser . .

CMD ["go", "test", "-v", "-cover", "./..."]
