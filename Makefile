VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git symbolic-ref -q --short HEAD)
COMMIT_HASH ?= $(shell git rev-parse --short HEAD 2>/dev/null)

DATE_FMT = +%FT%TZ # ISO 8601
BUILD_DATE ?= $(shell date "$(DATE_FMT)") # "-u" for UTC time (zero offset)

BUILD_DIR ?= bin
LDFLAGS += "-X main.version=$(VERSION) -X main.commitHash=$(COMMIT_HASH) -X main.buildDate=$(BUILD_DATE)"

.DEFAULT_GOAL: help
default: help

##@ Build

.PHONY: build
build: ## Build binaries.
	@mkdir -p $(BUILD_DIR)
	go build -ldflags $(LDFLAGS) -o $(BUILD_DIR)/ ./cmd/...

install:  ## Install binaries.
	go install -ldflags $(LDFLAGS) ./cmd/$* 

##@ Generate

gen: ## Generates code and documentation (see: ./gen.go).
	go generate ./...

##@ Test and Lint

.PHONY: test coverage
test: ## Test go code.
	go test -ldflags $(LDFLAGS) -v -cover -race ./...
coverage:  ## Test and check code coverage.
	go test -ldflags $(LDFLAGS) -short ./... -coverprofile cover.out 2>/dev/null
	go tool cover -func cover.out

.PHONY: lint
lint: ## See lint violations.
	golangci-lint run ./...

FORMATTING_BEGIN_YELLOW = \033[0;33m
FORMATTING_BEGIN_BLUE = \033[36m
FORMATTING_END = \033[0m

.PHONY: help
help:
	@printf -- "${FORMATTING_BEGIN_BLUE}%s${FORMATTING_END}\n" \
	"" \
	"     :?~             ^?:      											" \
	"   ^Y&@@P~         ~P@@&5^    											" \
	"  7@@@@@@@G!       J&@@@@@J   Omlox Hub™ go client library.			" \
	"   ~P@@@@@@@B7.     .?B@G7.   											" \
	"     ^5&@@@@@@#?.     .^      version: $(VERSION) ($(COMMIT_HASH)) 	" \
	"       ^Y&@@@@@@#J:           											" \
	"         :J#@@@@@@&Y^         											" \
	"           .?#@@@@@@&5^       											" \
	"    ~P?.     .7B@@@@@@@P~     											" \
	"  ~B@@@#J:      !G@@@@@@@B!   											" \
	"  ^5&@@@@P.       ~P@@@@@P~   											" \
	"    :J#P~           ^5#5^     											" \
	"      .               .	   											" \
	"" \
	"-----------------------------------------------------------------------" \
	""
	@awk 'BEGIN {\
	    FS = ":.*##"; \
	    printf                "Usage: ${FORMATTING_BEGIN_BLUE}OPTION${FORMATTING_END}=<value> make ${FORMATTING_BEGIN_YELLOW}<target>${FORMATTING_END}\n"\
	  } \
	  /^[a-zA-Z0-9_-]+:.*?##/ { printf "  ${FORMATTING_BEGIN_BLUE}%-36s${FORMATTING_END} %s\n", $$1, $$2 } \
	  /^.?.?##~/              { printf "   %-46s${FORMATTING_BEGIN_YELLOW}%-46s${FORMATTING_END}\n", "", substr($$1, 6) } \
	  /^##@/                  { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)