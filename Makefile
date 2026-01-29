PROTO_FILES := $(shell find api -name *.proto)

.PHONY: install
install:
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: conf
conf:
	cp config.example.yaml config.yaml

.PHONY: generate
generate:
	go generate ./...

lint:
	golangci-lint run --config ./.golangci.yml

test:
	go test -v ./...

.PHONY: run
run:
	@go run ./cmd/lumenim $(filter-out run,$(MAKECMDGOALS))

.PHONY: dev-http dev-comet dev-queue dev-crontab
dev-http:
	@go run ./cmd/lumenim http
dev-comet:
	@go run ./cmd/lumenim comet
dev-queue:
	@go run ./cmd/lumenim queue
dev-crontab:
	@go run ./cmd/lumenim crontab

.PHONY: dev
dev:
	@$(MAKE) -j 4 dev-http dev-comet dev-queue dev-crontab

.PHONY: build
build:
	go build -o ./bin/lumenim ./cmd/lumenim

.PHONY: protoc-gen-bff
protoc-gen-bff:
	go build -o protoc-gen-bff ./cmd/protoc-gen-bff

.PHONY: proto
proto: protoc-gen-bff
	@if [ -n "$(PROTO_FILES)" ]; then \
		protoc \
		--plugin=protoc-gen-bff=./protoc-gen-bff \
		--proto_path=./api/proto \
		--proto_path=./third_party \
		--go_out=paths=source_relative:./api/pb/ \
		--bff_out=./api/pb/ \
		$(PROTO_FILES) \
	 && echo "protoc generate success"; \
	fi
	make proto-openapi
	@if [ -f ./protoc-gen-bff ]; then \
    	rm -f ./protoc-gen-bff; \
    fi

.PHONY: proto-openapi # 生成 OpenApi 文档
proto-openapi:
	@for dir in $$(find api/proto -type d -mindepth 1 -maxdepth 1); do \
		echo "Processing directory: $$dir"; \
		proto_files=$$(find $$dir -name "*.proto" ! -name "article*"); \
		if [ -n "$$proto_files" ]; then \
		  protoc \
          			--proto_path=./api/proto \
          			--proto_path=./third_party \
          			--openapi_out=version=3:./$$dir \
          			$$proto_files; \
		fi; \
		echo "Generated OpenAPI spec for directory: $$dir";\
	done

SWAG_BIN := $(shell go env GOPATH)/bin/swag

.PHONY: swagger-install # 安装 Swag 工具
swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: swagger-gen # 生成 Swagger 文档
swagger-gen:
	$(SWAG_BIN) init -g internal/apis/server.go -o docs --parseDependency --parseInternal --exclude internal/apis/handler/web/v1/article

.PHONY: swagger-fmt # 格式化 Swagger 注释
swagger-fmt:
	$(SWAG_BIN) fmt -g internal/apis/server.go


## 自定义命令
-include custom.mk