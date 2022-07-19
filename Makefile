APPNAME := aws-billing-store
STAGE ?= dev
BRANCH ?= master

GOLANGCI_VERSION = v1.47.0

GIT_HASH := $(shell git rev-parse --short HEAD)

DEPLOY_CMD = sam deploy

.PHONY: ci
ci: lint test build

bin/golangci-lint: bin/golangci-lint-${GOLANGCI_VERSION}
	@ln -sf golangci-lint-${GOLANGCI_VERSION} bin/golangci-lint
bin/golangci-lint-${GOLANGCI_VERSION}:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_VERSION)
	@mv bin/golangci-lint $@

.PHONY: deploy
deploy: clean build archive deploy-cur-bucket deploy-cur deploy-symlink deploy-athena deploy-athena-workspace deploy-partitions

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

.PHONY: deploy-bucket
deploy-bucket:
	@echo "--- deploy stack deployment-$(STAGE)-$(BRANCH)"
	@$(DEPLOY_CMD) \
		--no-fail-on-empty-changeset \
		--template-file sam/app/symlink.yaml \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=deployment" \
		--stack-name deployment-$(STAGE)-$(BRANCH) \
		--parameter-overrides AppName=deployment Stage=$(STAGE) Branch=$(BRANCH)

.PHONY: deploy-cur-bucket
deploy-cur-bucket:
	@echo "--- deploy stack $(APPNAME)-$(STAGE)-$(BRANCH)-cur-bucket"
	$(eval SAM_BUCKET := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/deploy_bucket' --query 'Parameter.Value' --output text))

	@$(DEPLOY_CMD) \
		--no-fail-on-empty-changeset \
		--template-file sam/app/cur-bucket.yaml \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=$(APPNAME)" \
		--stack-name $(APPNAME)-$(STAGE)-$(BRANCH)-cur-bucket \
		--parameter-overrides AppName=$(APPNAME) Stage=$(STAGE) Branch=$(BRANCH) ReportPrefix=cur/$(APPNAME)-$(STAGE)-$(BRANCH)-cur-$(AWS_DEFAULT_REGION)-athena-hourly

.PHONY: deploy-cur
deploy-cur:
	@echo "--- deploy stack $(APPNAME)-$(STAGE)-$(BRANCH)-cur"
	$(eval SAM_BUCKET := $(shell aws ssm --region us-east-1 get-parameter --name '/config/$(STAGE)/$(BRANCH)/deploy_bucket' --query 'Parameter.Value' --output text))
	$(eval CUR_BUCKET_NAME := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/report_bucket' --query 'Parameter.Value' --output text))
	$(eval CUR_PREFIX := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/report_prefix' --query 'Parameter.Value' --output text))

	@$(DEPLOY_CMD) \
		--no-fail-on-empty-changeset \
		--region us-east-1 \
		--template-file sam/app/cur.yaml \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=$(APPNAME)" \
		--stack-name $(APPNAME)-$(STAGE)-$(BRANCH)-cur-$(AWS_DEFAULT_REGION) \
		--parameter-overrides AppName=$(APPNAME) Stage=$(STAGE) Branch=$(BRANCH) \
			ReportBucketName=$(CUR_BUCKET_NAME) ReportBucketRegion=$(AWS_DEFAULT_REGION)

.PHONY: deploy-symlink
deploy-symlink:
	@echo "--- deploy stack $(APPNAME)-$(STAGE)-$(BRANCH)-symlink"
	$(eval SAM_BUCKET := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/deploy_bucket' --query 'Parameter.Value' --output text))
	$(eval CUR_BUCKET_NAME := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/report_bucket' --query 'Parameter.Value' --output text))
	$(eval CUR_PREFIX := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/report_prefix' --query 'Parameter.Value' --output text))

	@$(DEPLOY_CMD) \
		--no-fail-on-empty-changeset \
		--template-file sam/app/symlink.yaml \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=$(APPNAME)" \
		--stack-name $(APPNAME)-$(STAGE)-$(BRANCH)-symlink \
		--parameter-overrides AppName=$(APPNAME) Stage=$(STAGE) Branch=$(BRANCH) Commit=$(GIT_HASH) \
			ReportBucketName=$(CUR_BUCKET_NAME) CurPrefix=$(CUR_PREFIX)

.PHONY: deploy-athena-workspace
deploy-athena-workspace:
	@echo "--- deploy stack $(APPNAME)-$(STAGE)-$(BRANCH)-athena-workspace"
	$(eval SAM_BUCKET := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/deploy_bucket' --query 'Parameter.Value' --output text))

	@$(DEPLOY_CMD) \
		--no-fail-on-empty-changeset \
		--template-file sam/app/athena-workspace.yaml \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=$(APPNAME)" \
		--stack-name $(APPNAME)-$(STAGE)-$(BRANCH)-athena-workspace \
		--parameter-overrides AppName=$(APPNAME) Stage=$(STAGE) Branch=$(BRANCH) Commit=$(GIT_HASH)

.PHONY: deploy-athena
deploy-athena:
	@echo "--- deploy stack $(APPNAME)-$(STAGE)-$(BRANCH)-athena"
	$(eval SAM_BUCKET := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/deploy_bucket' --query 'Parameter.Value' --output text))
	$(eval CUR_BUCKET_NAME := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/report_bucket' --query 'Parameter.Value' --output text))
	$(eval CUR_PREFIX := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/report_prefix' --query 'Parameter.Value' --output text))

	@$(DEPLOY_CMD) \
		--no-fail-on-empty-changeset \
		--template-file sam/app/athena.yaml \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=$(APPNAME)" \
		--stack-name $(APPNAME)-$(STAGE)-$(BRANCH)-athena \
		--parameter-overrides AppName=$(APPNAME) Stage=$(STAGE) Branch=$(BRANCH) Commit=$(GIT_HASH) \
			ReportBucketName=$(CUR_BUCKET_NAME) CurPrefix=$(CUR_PREFIX)

.PHONY: deploy-partitions
deploy-partitions:
	@echo "--- deploy stack $(APPNAME)-$(STAGE)-$(BRANCH)-partitions"
	$(eval SAM_BUCKET := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/deploy_bucket' --query 'Parameter.Value' --output text))
	$(eval CUR_BUCKET_NAME := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/report_bucket' --query 'Parameter.Value' --output text))
	$(eval CUR_PREFIX := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/report_prefix' --query 'Parameter.Value' --output text))
	$(eval QUERY_RESULTS_BUCKET_NAME := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/athena_query_results_bucketname' --query 'Parameter.Value' --output text))
	$(eval GLUE_DATABASE_NAME := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/glue_database_name' --query 'Parameter.Value' --output text))
	$(eval GLUE_TABLE_NAME := $(shell aws ssm get-parameter --name '/config/$(STAGE)/$(BRANCH)/$(APPNAME)/glue_table_name' --query 'Parameter.Value' --output text))
	@$(DEPLOY_CMD) \
		--no-fail-on-empty-changeset \
		--template-file sam/app/partitions.yaml \
		--capabilities CAPABILITY_IAM \
		--s3-bucket $(SAM_BUCKET) \
		--s3-prefix sam/$(GIT_HASH) \
		--tags "environment=$(STAGE)" "branch=$(BRANCH)" "service=$(APPNAME)" \
		--stack-name $(APPNAME)-$(STAGE)-$(BRANCH)-partitions \
		--parameter-overrides AppName=$(APPNAME) Stage=$(STAGE) Branch=$(BRANCH) Commit=$(GIT_HASH) \
			ReportBucketName=$(CUR_BUCKET_NAME) \
			CurPrefix=$(CUR_PREFIX) \
			QueryResultsBucketName=$(QUERY_RESULTS_BUCKET_NAME) \
			GlueDatabase=$(GLUE_DATABASE_NAME) \
			GlueTable=$(GLUE_TABLE_NAME)
