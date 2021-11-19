FROM archlinux:base-20201220.0.11678

WORKDIR /app

RUN pacman -Sy -q --noconfirm \
  glibc \
  git \
  gcc \
  go

RUN useradd -m someuser

USER someuser

COPY --chown=someuser:someuser . .

CMD ["go", "test", "-v", "-cover", "./..."]
