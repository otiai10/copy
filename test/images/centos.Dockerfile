FROM centos:centos8

WORKDIR /app

RUN yum update -y -q \
    && yum install -y --quiet \
      git \
      go

RUN useradd -m someuser

USER someuser

COPY --chown=someuser:someuser . .

CMD ["go", "test", "-v", "-cover", "./..."]
