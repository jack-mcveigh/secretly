all: build

build:
	go build -o bin/ ./cmd/...

clean:
	$(RM) -r bin *.out

coverage-report:
	go test -coverprofile cover.out ./...

coverage-report-visual: coverage-report
	go tool cover -html=cover.out

test unit-test:
	go test -v -cover ./...

.PHONY: all build clean coverage-report coverage-report-visual test unit-test
