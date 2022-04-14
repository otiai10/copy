FROM centos:centos8

WORKDIR /app

RUN cd /etc/yum.repos.d/ \
  && sed -i 's/mirrorlist/#mirrorlist/g' /etc/yum.repos.d/CentOS-* \
  && sed -i 's|#baseurl=http://mirror.centos.org|baseurl=http://vault.centos.org|g' /etc/yum.repos.d/CentOS-*
RUN dnf upgrade -y --quiet
RUN dnf install -y --quiet git go

RUN useradd -m someuser

# root-owned directory setup for test case 14
RUN mkdir -p test/owned-by-root \
  && chown :someuser test/owned-by-root \
  && chmod 775 test/owned-by-root

USER someuser

COPY --chown=someuser:someuser . .

CMD ["go", "test", "-v", "-cover", "./..."]
