# syntax=docker/dockerfile:1

ARG POSTGRES_VERSION

FROM postgres:${POSTGRES_VERSION} AS production

ARG POSTGRES_VERSION
ARG DOCUMENTDB_VERSION

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

# Remove verification after https://github.com/pgaudit/pgaudit/issues/257 is resolved and pgaudit builds with PG17
if [ "${POSTGRES_VERSION}" -lt 17 ]; then
    apt install -y postgresql-${POSTGRES_VERSION}-pgaudit
fi

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
LABEL org.opencontainers.image.source="https://github.com/FerretDB/documentdb"
LABEL org.opencontainers.image.url="https://www.ferretdb.com/"
LABEL org.opencontainers.image.vendor="FerretDB Inc."
