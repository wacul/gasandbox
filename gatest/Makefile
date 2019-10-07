build:
	GO111MODULE=on go build -o dist/ga ./main.go

N ?= 200
I ?= 0.2
S ?= secret.json
C ?= 0
start:
	${MAKE} build
	./dist/ga --secret=${S} --count=${N} --concurrent=${C} --interval=${I}

docker_run:
	docker build -t gatest .
	docker run --rm gatest ga --secret=${S} --count=${N} --concurrent=${C}

.PHONY: build
.PHONY: start
.PHONY: docker_run
