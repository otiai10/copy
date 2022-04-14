FROM archlinux:base-20201220.0.11678

WORKDIR /app

RUN pacman -Sy -q --noconfirm archlinux-keyring
RUN pacman -Sy -q --noconfirm \
  glibc \
  git \
  gcc \
  go

RUN useradd -m someuser

# root-owned directory setup for test case 14
RUN mkdir -p test/owned-by-root \
  && chown :someuser test/owned-by-root \
  && chmod 775 test/owned-by-root

USER someuser

COPY --chown=someuser:someuser . .

CMD ["go", "test", "-v", "-cover", "./..."]
