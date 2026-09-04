GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_NAME=api-mock-service
BRANCH=$(shell git rev-parse --symbolic-full-name --abbrev-ref HEAD)
COMMIT?=$(shell git describe --always --long --dirty)
VERSION?=$(shell git describe --always --long --dirty)
DATE?=$(shell date -u '+%Y-%m-%dT%H:%M:%S')
SERVICE_PORT?=3000
TEST_RACE_PROCESS=-race
PKG_LIST=$(shell go list ./... | grep -v /vendor/)
EXPORT_RESULT?=false # for CI please set EXPORT_RESULT to true

all: test vendor build

build: vendor
	mkdir -p out/bin
	GO111MODULE=on $(GOCMD) build -mod vendor -ldflags "-X main.commit=$(COMMIT) -X main.date=$(DATE) -X main.version=$(VERSION)" -o out/bin/$(BINARY_NAME) .

build-linux: vendor
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOCMD) build -mod=vendor -o "out/bin/$(BINARY_NAME)_{{.OS}}_{{.Arch}}" -v

clean:
	rm -fr ./bin
	rm -fr ./out
	rm -fr ./vendor
	rm -f ./junit-report.xml checkstyle-report.xml ./coverage.xml ./profile.cov yamllint-checkstyle.xml

coverage:
	$(GOTEST) -cover -covermode=count -coverprofile=profile.cov ./...
	$(GOCMD) tool cover -func profile.cov
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/AlekSi/gocov-xml
	GO111MODULE=off go get -u github.com/axw/gocov/gocov
	gocov convert profile.cov | gocov-xml > coverage.xml
endif


SWAGGER=$(shell go env GOPATH)/bin/swagger

# npm install -g swagger-markdown
check-swagger:
	@test -x $(SWAGGER) || go install github.com/go-swagger/go-swagger/cmd/swagger@latest

swagger: check-swagger
	GO111MODULE=on go mod vendor && $(SWAGGER) generate spec -o ./docs/swagger.yaml && $(SWAGGER) generate spec -o ./docs/swagger.json
	which swagger-markdown > /dev/null 2>&1 && swagger-markdown -i docs/swagger.yaml || true

serve-swagger: swagger
	$(SWAGGER) serve -F=swagger docs/swagger.yaml

docker-build:
	docker build --rm --tag $(BINARY_NAME) .

docker-release:
	docker tag $(BINARY_NAME) plexobject/$(DOCKER_REGISTRY)$(BINARY_NAME):latest
	#docker tag $(BINARY_NAME) plexobject/$(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)
	# Push the docker images
	docker push plexobject/$(DOCKER_REGISTRY)$(BINARY_NAME):latest
	#docker push plexobject/$(DOCKER_REGISTRY)$(BINARY_NAME):$(VERSION)

lint: 
	golangci-lint run --enable-all

vet: clean 
	$(GOVET) ./... 2> go-vet-report.out

run: build
	./"out/bin/${BINARY_NAME}"

test:
ifeq ($(EXPORT_RESULT), true)
	GO111MODULE=off go get -u github.com/jstemmer/go-junit-report
	$(eval OUTPUT_OPTIONS = | tee /dev/tty | go-junit-report -set-exit-code > junit-report.xml)
endif
	@rm -rf mock_tests/api_contracts/2 mock_tests/api_contracts/v1 mock_tests/api_contracts/v2 mock_tests/api_contracts/v3
	$(GOTEST) -count=1 ./... $(OUTPUT_OPTIONS)

test-race:
	$(GOTEST) -v -race -count=1 ./...

vendor:
	$(GOCMD) mod vendor

.PHONY: vendor build test

