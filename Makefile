VERSION ?= $(shell git describe --always --dirty 2>/dev/null || echo dev)
APP_PACKAGE := github.com/yaqyn/readlane/internal/app

.PHONY: build install test integration vet generate check-generated audit verify smoke
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X $(APP_PACKAGE).Version=$(VERSION)" -o readlane ./cmd/readlane
install: build
	install -Dm755 readlane "$(HOME)/.local/bin/readlane"
test:
	go test -race ./...
integration:
	@test -n "$$READLANE_TEST_DATABASE_URL" || (echo 'Set READLANE_TEST_DATABASE_URL to a disposable database ending in _test' >&2; exit 1)
	go test -race ./...
smoke: build
	python3 scripts/verify-binary.py ./readlane
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
