VERSION ?= $(shell git describe --always --dirty 2>/dev/null || echo dev)
APP_PACKAGE := github.com/yaqyn/blog-aggregator/internal/app

.PHONY: build install test integration vet generate check-generated audit verify smoke
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X $(APP_PACKAGE).Version=$(VERSION)" -o gator ./cmd/gator
install: build
	install -Dm755 gator "$(HOME)/.local/bin/gator"
test:
	go test -race ./...
integration:
	@test -n "$$GATOR_TEST_DATABASE_URL" || (echo 'Set GATOR_TEST_DATABASE_URL to a disposable database ending in _test' >&2; exit 1)
	go test -race ./...
smoke: build
	python3 scripts/verify-binary.py ./gator
vet:
	go vet ./...
generate:
	sqlc generate
check-generated: generate
	git diff --exit-code -- internal/database
	@test -z "$$(git ls-files --others --exclude-standard internal/database)"
audit:
	govulncheck ./...
verify: vet test check-generated audit
