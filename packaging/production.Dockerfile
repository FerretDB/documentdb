# syntax=docker/dockerfile:1

ARG POSTGRES_VERSION

FROM postgres:${POSTGRES_VERSION}

ARG TARGETARCH
ARG POSTGRES_VERSION
ARG DOCUMENTDB_VERSION

# common steps for production and development

RUN --mount=type=cache,sharing=locked,target=/var/cache/apt <<EOF
set -ex

apt update
apt upgrade -y
apt install -y \
    postgresql-${POSTGRES_VERSION} \
    postgresql-${POSTGRES_VERSION}-cron \
    postgresql-${POSTGRES_VERSION}-pgvector \
    postgresql-${POSTGRES_VERSION}-postgis-3 \
    postgresql-${POSTGRES_VERSION}-rum \
    postgresql-server-dev-${POSTGRES_VERSION} \
    barman-cli-cloud
EOF

RUN --mount=target=/src,rw <<EOF
set -ex

cd /src

cp packaging/deb12-postgresql-${POSTGRES_VERSION}-documentdb_${DOCUMENTDB_VERSION}_${TARGETARCH}.deb /tmp/documentdb.deb
dpkg -i /tmp/documentdb.deb
rm /tmp/documentdb.deb

cp packaging/10-preload.sh  /docker-entrypoint-initdb.d/
cp packaging/20-install.sql /docker-entrypoint-initdb.d/

EOF

WORKDIR /

LABEL org.opencontainers.image.title="PostgreSQL+DocumentDB"
LABEL org.opencontainers.image.description="PostgreSQL with DocumentDB extension"
LABEL org.opencontainers.image.source="https://github.com/FerretDB/documentdb"
LABEL org.opencontainers.image.url="https://www.ferretdb.com/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
