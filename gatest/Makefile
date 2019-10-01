build:
	GO111MODULE=on go build -o dist/ga ./main.go

N ?= 200
I ?= 0.2
S ?= secret.json
start:
	${MAKE} build
	./dist/ga --secret=${S} --count=${N} --interval=${I}

docker_run:
	docker build -t gatest .
	docker run --rm gatest ga --secret=${S} --count=${N} --interval=${I}

.PHONY: build
.PHONY: start
.PHONY: docker_run
