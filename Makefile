all: build-examples

build-examples:
	go build -o examples/bin/ ./examples/...

clean:
	@$(RM) -r examples/bin *.out

coverage-report:
	go test -coverprofile cover.out ./...

coverage-report-visual: coverage-report
	go tool cover -html=cover.out

test unit-test: 
	go test -v -cover ./...

.PHONY: all build-examples clean
.PHONY: coverage-report coverage-report-visual test unit-test