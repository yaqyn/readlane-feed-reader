VERSION ?= $(shell git describe --always --dirty 2>/dev/null || echo dev)
APP_PACKAGE := github.com/yaqyn/readlane-feed-reader/internal/app

.PHONY: build install test integration vet generate check-generated audit verify smoke web-check design-check startup-check
build:
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X $(APP_PACKAGE).Version=$(VERSION)" -o readlane ./cmd/readlane
install: build
	install -Dm755 readlane "$(HOME)/.local/bin/readlane"
startup-check: build
	python3 scripts/verify-startup.py ./readlane ./readlane ./launch
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
web-check:
	npm run test:web
	npm exec -- prettier --check internal/web/assets/reader.html internal/web/assets/reader.css internal/web/assets/reader.js tests/reader.test.cjs package.json premium-ui.json
design-check:
	python3 scripts/design-tokens.py --check
verify: vet test check-generated design-check audit
