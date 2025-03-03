.PHONY: docker-init docker-build docker-cleanup

docker-init:
	docker buildx create \
		--driver=docker-container \
		--name=documentdb \
		--bootstrap=true \
		--use=false \
		--config=./buildkitd.toml \
		--driver-opt network=host \
		--driver-opt env.JAEGER_TRACE=127.0.0.1:6831 \
		--driver-opt env.BUILDKIT_STEP_LOG_MAX_SIZE=-1 \
		--driver-opt env.BUILDKIT_STEP_LOG_MAX_SPEED=-1 \
		|| true
	docker buildx ls

docker-build:
	test -n "$(POSTGRES_VERSION)" || (echo "POSTGRES_VERSION not set" && false)
	test -n "$(DOCUMENTDB_VERSION)" || (echo "DOCUMENTDB_VERSION not set" && false)
	test -n "$(FILE)" || (echo "FILE not set" && false)
	test -n "$(OUTPUT)" || (echo "OUTPUT not set" && false)
	test -n "$(TAGS)" || (echo "TAGS not set" && false)
	docker buildx build --builder=documentdb \
		--file=ferretdb_packaging/$(TARGET).Dockerfile \
		--build-arg='POSTGRES_VERSION=$(POSTGRES_VERSION)' \
		--build-arg='DOCUMENTDB_VERSION=$(DOCUMENTDB_VERSION)' \
		--platform=linux/amd64 \
		--output='$(OUTPUT)' \
		$(TAGS) \
		..

docker-cleanup:
	docker system df
	docker buildx --builder=documentdb du || true
	docker buildx --builder=documentdb rm --force || true
	docker system prune --force
	docker system df
