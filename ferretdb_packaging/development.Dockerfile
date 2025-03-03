# syntax=docker/dockerfile:1

ARG PG_MAJOR=16

FROM postgres:${PG_MAJOR} AS development

ARG DOCUMENTDB_VERSION

RUN --mount=type=cache,sharing=locked,target=/var/cache/apt <<EOF
set -ex

apt update
apt upgrade -y
apt install -y \
    postgresql-${PG_MAJOR}-cron \
    postgresql-${PG_MAJOR}-pgvector \
    postgresql-${PG_MAJOR}-postgis-3 \
    postgresql-${PG_MAJOR}-rum \
EOF

RUN --mount=target=/src,rw <<EOF
set -ex

cd /src

cp packaging/deb12-postgresql-${PG_MAJOR}-documentdb_${DOCUMENTDB_VERSION}_amd64.deb /tmp/documentdb.deb
dpkg -i /tmp/documentdb.deb
rm /tmp/documentdb.deb

EOF

# extra packages for development
RUN --mount=type=cache,sharing=locked,target=/var/cache/apt <<EOF
set -ex

apt install -y \
    postgresql-${PG_MAJOR}-pgtap
EOF

RUN --mount=target=/src,rw <<EOF
set -ex

cd /src

cp ferretdb_packaging/10-preload.sh ferretdb_packaging/20-install.sql /docker-entrypoint-initdb.d/
cp ferretdb_packaging/90-install-development.sql /docker-entrypoint-initdb.d/

EOF

WORKDIR /

LABEL org.opencontainers.image.title="PostgreSQL+DocumentDB (development image)"
LABEL org.opencontainers.image.description="PostgreSQL with DocumentDB extension (development image)"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/FerretDB"
LABEL org.opencontainers.image.url="https://www.ferretdb.com/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
