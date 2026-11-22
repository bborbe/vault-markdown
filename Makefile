BRANCH ?= $(shell git rev-parse --abbrev-ref HEAD)

include tools.env

.PHONY: default
default: precommit

.PHONY: precommit
precommit: ensure format test check addlicense
	@echo "ready to commit"

.PHONY: ensure
ensure:
	go mod tidy -e
	go mod verify
	rm -rf vendor

.PHONY: format
format:
	find . -type f -name 'go.mod' -not -path './vendor/*' -exec go run github.com/shoenig/go-modtool@$(GO_MODTOOL_VERSION) -w fmt "{}" \;
	go run github.com/incu6us/goimports-reviser/v3@$(GOIMPORTS_REVISER_VERSION) -project-name github.com/bborbe/vault-markdown -format -excludes vendor ./...
	find . -type d -name vendor -prune -o -type f -name '*.go' -print0 | xargs -0 -n 10 go run github.com/segmentio/golines@$(GOLINES_VERSION) --max-len=100 -w
	# golines last, then gofmt last so its wrapping is normalized and the gofmt lint check passes
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +

.PHONY: test
test:
	go test -mod=mod -p=$${GO_TEST_PARALLEL:-1} -cover -race=false $(shell go list -mod=mod ./... | grep -v /vendor/)

.PHONY: check
check: lint vet vulncheck osv-scanner trivy

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run --allow-parallel-runners --config .golangci.yml ./...

.PHONY: vet
vet:
	go vet -mod=mod $(shell go list -mod=mod ./... | grep -v /vendor/)

.PHONY: vulncheck
vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) $(shell go list -mod=mod ./... | grep -v /vendor/)

.PHONY: osv-scanner
osv-scanner:
	go run github.com/google/osv-scanner/v2/cmd/osv-scanner@$(OSV_SCANNER_VERSION) --recursive .

.PHONY: trivy
trivy:
	trivy fs \
	--db-repository ghcr.io/aquasecurity/trivy-db \
	--scanners vuln,secret \
	--quiet \
	--no-progress \
	--disable-telemetry \
	--exit-code 1 .

.PHONY: addlicense
addlicense:
	go run github.com/google/addlicense@$(ADDLICENSE_VERSION) -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')
