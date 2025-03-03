.PHONY: docker-build docker-cleanup docker-documentdb docker-documentdb-development-push docker-documentdb-production-push docker-init

docker-build:
	test -n "$(POSTGRES_VERSION)" || (echo "POSTGRES_VERSION not set" && false)
	test -n "$(PG_MAJOR)" || (echo "PG_MAJOR not set" && false)
	test -n "$(DOCUMENTDB_VERSION)" || (echo "DOCUMENTDB_VERSION not set" && false)
	test -n "$(TARGET)" || (echo "TARGET not set" && false)
	test -n "$(OUTPUT)" || (echo "OUTPUT not set" && false)
	test -n "$(PLATFORM)" || (echo "PLATFORM not set" && false)
	test -n "$(TAGS)" || (echo "TAGS not set" && false)
	cd ..; docker buildx build --builder=ferretdb \
        --file=ferretdb_packaging/$(TARGET).Dockerfile \
        --build-arg=POSTGRES_VERSION=$(POSTGRES_VERSION) \
        --build-arg=PG_MAJOR=$(PG_MAJOR) \
        --build-arg=DOCUMENTDB_VERSION=$(DOCUMENTDB_VERSION) \
        --target=$(TARGET) \
        --platform=$(PLATFORM) \
        $(TAGS) \
        --output=$(OUTPUT) \
        .

docker-cleanup:
	docker system df
	docker buildx --builder=ferretdb du || true # ignore error
	docker buildx --builder=ferretdb rm --force || true # ignore error
	docker system prune --force
	docker system df

docker-documentdb:
	test -n "$(POSTGRES_VERSION)" || (echo "POSTGRES_VERSION not set" && false)
	test -n "$(PG_MAJOR)" || (echo "PG_MAJOR not set" && false)
	test -n "$(DOCUMENTDB_VERSION)" || (echo "DOCUMENTDB_VERSION not set" && false)
	test -n "$(TARGET)" || (echo "TARGET not set" && false)
	make docker-build \
        POSTGRES_VERSION=$(POSTGRES_VERSION) \
        PG_MAJOR=$(PG_MAJOR) \
        DOCUMENTDB_VERSION=$(DOCUMENTDB_VERSION) \
        TARGET=$(TARGET) \
        PLATFORM=linux/amd64 \
        TAGS="--tag=documentdb-local" \
        OUTPUT="type=docker"

docker-documentdb-development-push:
	test -n "$(POSTGRES_VERSION)" || (echo "POSTGRES_VERSION not set" && false)
	test -n "$(PG_MAJOR)" || (echo "PG_MAJOR not set" && false)
	test -n "$(DOCUMENTDB_VERSION)" || (echo "DOCUMENTDB_VERSION not set" && false)
	test -n "$(TAGS)" || (echo "TAGS not set" && false)
	make docker-build \
        POSTGRES_VERSION=$(POSTGRES_VERSION) \
        PG_MAJOR=$(PG_MAJOR) \
        DOCUMENTDB_VERSION=$(DOCUMENTDB_VERSION) \
        TARGET=development \
        PLATFORM=linux/amd64 \
        TAGS=$(TAGS)
        OUTPUT="type=image,push=true"

docker-documentdb-production-push:
	test -n "$(POSTGRES_VERSION)" || (echo "POSTGRES_VERSION not set" && false)
	test -n "$(PG_MAJOR)" || (echo "PG_MAJOR not set" && false)
	test -n "$(DOCUMENTDB_VERSION)" || (echo "DOCUMENTDB_VERSION not set" && false)
	test -n "$(TAGS)" || (echo "TAGS not set" && false)
	make docker-build \
        POSTGRES_VERSION=$(POSTGRES_VERSION) \
        PG_MAJOR=$(PG_MAJOR) \
        DOCUMENTDB_VERSION=$(DOCUMENTDB_VERSION) \
        TARGET=production \
        PLATFORM=linux/amd64 \
        TAGS=$(TAGS)
        OUTPUT="type=image,push=true"

docker-init:
	docker buildx create \
        --driver=docker-container \
        --name=ferretdb \
        --bootstrap=true \
        --use=false \
        --config=./buildkitd.toml \
        --driver-opt network=host \
        --driver-opt env.JAEGER_TRACE=127.0.0.1:6831 \
        --driver-opt env.BUILDKIT_STEP_LOG_MAX_SIZE=-1 \
        --driver-opt env.BUILDKIT_STEP_LOG_MAX_SPEED=-1 \
        || true # ignore error
	docker buildx ls

