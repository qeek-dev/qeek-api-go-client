REPO_OWNER := cage1016
PROHECT_NAME := qeek-api-go-client
COVERAGE_PATH = $(CURDIR)/bin/coverage
GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)

dep:
	@dep init

test:
	# DEBUG=true bash -c "go test -v github.com/qeek-dev/qeek-api-go-client/<package-name> -run ..."
	@go test -ldflags -s -v $(GOPACKAGES)

bench:
	# DEBUG=false bash -c "go test -v github.com/qeek-dev/qeek-api-go-client/route/routelogin -bench=. -run BenchmarkLoginHandler"
	@go test -v $(GOPACKAGES) -bench . -run=^Benchmark

coverage:
	@mkdir -p $(CURDIR)/${COVERAGE_PATH}
	@docker run --rm -v ${PWD}:/go/src/github.com/$(REPO_OWNER)/$(PROHECT_NAME) -w /go/src/github.com/$(REPO_OWNER)/$(PROHECT_NAME) garychen/golang:1.10-alpine bash -c \
	'	gocov test ./... > $${PWD}/$(COVERAGE_PATH)/coverage.out && \
		gocov report $${PWD}/$(COVERAGE_PATH)/coverage.out && \
		if test -z "$$CI"; then \
			gocov-html $${PWD}/$(COVERAGE_PATH)/coverage.out > $${PWD}/$(COVERAGE_PATH)/coverage.html; \
		fi'
	@open $(CURDIR)/${COVERAGE_PATH}/coverage.html


.PHONY: dep run test bench coverage
