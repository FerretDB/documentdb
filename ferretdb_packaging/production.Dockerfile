# syntax=docker/dockerfile:1

ARG POSTGRES_VERSION=16

FROM postgres:${POSTGRES_VERSION} AS production

ARG DOCUMENTDB_VERSION

# redeclare arg to use within build
# https://docs.docker.com/reference/dockerfile/#understand-how-arg-and-from-interact
ARG POSTGRES_VERSION

RUN --mount=type=cache,sharing=locked,target=/var/cache/apt <<EOF
set -ex

apt update
apt upgrade -y
apt install -y \
    postgresql-${POSTGRES_VERSION}-cron \
    postgresql-${POSTGRES_VERSION}-pgvector \
    postgresql-${POSTGRES_VERSION}-postgis-3 \
    postgresql-${POSTGRES_VERSION}-rum \
EOF

RUN --mount=target=/src,rw <<EOF
set -ex

cd /src

cp packaging/deb12-postgresql-${POSTGRES_VERSION}-documentdb_${DOCUMENTDB_VERSION}_amd64.deb /tmp/documentdb.deb
dpkg -i /tmp/documentdb.deb
rm /tmp/documentdb.deb

EOF

RUN --mount=target=/src,rw <<EOF
set -ex

cd /src

cp ferretdb_packaging/10-preload.sh ferretdb_packaging/20-install.sql /docker-entrypoint-initdb.d/

EOF

WORKDIR /

LABEL org.opencontainers.image.title="PostgreSQL+DocumentDB"
LABEL org.opencontainers.image.description="PostgreSQL with DocumentDB extension"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/FerretDB"
LABEL org.opencontainers.image.url="https://www.ferretdb.com/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
