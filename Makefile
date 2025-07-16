all:
	@echo Make targets are 'benckmarks', 'fmt', 'tests'

.PHONY: fmt
fmt:
	gofmt -s -w .

.PHONY: test tests
test tests:
	go test ./...
	go vet ./...

.PHONY: benchmark benchmarks
benchmark benchmarks:
	go test -benchmem -bench Bench
