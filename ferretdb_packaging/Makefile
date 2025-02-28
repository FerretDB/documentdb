.PHONY: docker-build docker-cleanup docker-documentdb docker-documentdb-development-push docker-documentdb-production-push docker-init

docker-build:
	test -n "$(FILE)" || (echo "FILE not set" && false)
	test -n "$(TARGET)" || (echo "TARGET not set" && false)
	test -n "$(OUTPUT)" || (echo "OUTPUT not set" && false)
	test -n "$(PLATFORM)" || (echo "PLATFORM not set" && false)
	test -n "$(TAGS)" || (echo "TAGS not set" && false)
	cd ..; docker buildx build --builder=ferretdb \
        --file=ferretdb_packaging/$(FILE).Dockerfile \
        --build-arg=LABEL_VERSION={{.VERSION}} \
        --build-arg=LABEL_COMMIT={{.COMMIT}} \
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
	test -n "$(FILE)" || (echo "FILE not set" && false)
	make docker-build \
        FILE=$(FILE) \
        TARGET=$(FILE) \
        PLATFORM=linux/amd64 \
        TAGS="--tag=documentdb-local" \
        OUTPUT="type=docker"

docker-documentdb-development-push:
	test -n "$(TAGS)" || (echo "TAGS not set" && false)
	make docker-build \
        FILE=development \
        TARGET=development \
        PLATFORM=linux/amd64 \
        OUTPUT="type=image,push=true"

docker-documentdb-production-push:
	test -n "$(TAGS)" || (echo "TAGS not set" && false)
	make docker-build \
        FILE=production \
        TARGET=production \
        PLATFORM=linux/amd64 \
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

