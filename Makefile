APPNAME := aws-billing-service
STAGE ?= dev
BRANCH ?= master

GOLANGCI_VERSION = v1.46.2

GIT_HASH := $(shell git rev-parse --short HEAD)

.PHONY: ci
ci: test build

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.46.2
	@mv bin/golangci-lint $@

.PHONY: lint
lint: bin/golangci-lint
	@echo "--- lint all the things"
	@bin/golangci-lint run

.PHONY: test
test:
	@go test -coverprofile=coverage.txt ./... > /dev/null
	@go tool cover -func=coverage.txt

.PHONY: build
build:
	CGO_ENABLED=0 GOAMD64=v2 go build -ldflags "-s -w -X main.commit=$(GIT_HASH)" -o dist/ ./cmd/...

.PHONY: clean
clean:
	rm -rf dist

.PHONY: archive
archive:
	@echo "--- build an archive"
	@cd dist && zip -X -9 ./handler.zip *-lambda

.PHONY: deploy-symlink
deploy-symlink:
	@echo "--- deploy stack $(APPNAME)-$(STAGE)-$(BRANCH)-symlink"
	$(eval SAM_BUCKET := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/deploy_bucket' --query 'Parameter.Value' --output text))

	@sam deploy \
		--no-fail-on-empty-changeset \
		--template-file sam/app/symlink.yaml \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=$(APPNAME)" \
		--stack-name $(APPNAME)-$(STAGE)-$(BRANCH)-symlink \
		--parameter-overrides AppName=$(APPNAME) Stage=$(STAGE) Branch=$(BRANCH) Commit=$(GIT_HASH) \
			DataBucketName=$(DATA_BUCKET_NAME) CurPrefix=$(CUR_PREFIX)