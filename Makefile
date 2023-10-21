all:
	@echo Make targets are 'fmt' and 'tests'

.PHONY: fmt
fmt:
	find . -name '*.go' -type f -print | xargs gofmt -s -w

.PHONY: test tests
test tests:
	go test ./...
	go vet ./...

.PHONY: benchmark benchmarks
benchmark benchmarks:
#	-benchtime 3s
	go test -benchmem -bench Bench
