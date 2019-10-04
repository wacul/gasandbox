build:
	GO111MODULE=on go build -o dist/ga ./main.go

N ?= 200
S ?= secret.json
C ?= 5
start:
	${MAKE} build
	./dist/ga --secret=${S} --count=${N} --concurrent=${C}

docker_run:
	docker build -t gatest .
	docker run --rm gatest ga --secret=${S} --count=${N} --concurrent=${C}

.PHONY: build
.PHONY: start
.PHONY: docker_run
