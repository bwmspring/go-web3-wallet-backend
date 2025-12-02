# 应用基本信息
APP_NAME        := wallet-backend
PROJECT_MODULE  := github.com/bwmspring/go-web3-wallet-backend
VERSION         ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.0.0-dev")
BUILD_TIME      := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 目录配置
BUILD_DIR       := ./bin
MAIN_PACKAGE    := ./main.go
CMD_PATH        := ./cmd
SRC_DIRS        := .

# 编译目标平台
GOOS            ?= $(shell go env GOOS)
GOARCH          ?= $(shell go env GOARCH)
TARGET          := $(BUILD_DIR)/$(APP_NAME)

# 构建标志
LDFLAGS         := -ldflags "-s -w \
                   -X $(PROJECT_MODULE)/cmd.Version=$(VERSION) \
                   -X $(PROJECT_MODULE)/cmd.BuildTime=$(BUILD_TIME) \
                   -X $(PROJECT_MODULE)/cmd.GitCommit=$(GIT_COMMIT)"

# 工具命令
GOLINES_CMD             := golines -w --max-len=120 --reformat-tags --shorten-comments --ignore-generated
GOIMPORTS_REVISER_CMD   := goimports-reviser -format -rm-unused -local $(PROJECT_MODULE)
GOLANGCI_LINT           := golangci-lint

# 测试配置
TEST_TIMEOUT    := 10m
COVERAGE_FILE   := coverage.out

# 颜色输出
BLUE            := \033[34m
GREEN           := \033[32m
YELLOW          := \033[33m
RED             := \033[31m
RESET           := \033[0m


.PHONY: all help
.PHONY: build build-linux build-darwin
.PHONY: run dev
.PHONY: clean clean-all
.PHONY: test test-coverage test-race test-benchmark
.PHONY: lint lint-fix format fmt
.PHONY: deps-tidy
.PHONY: install-tools

all: help

## help: 显示帮助信息
help:
	@printf "$(GREEN)===========================================$(RESET)\n"
	@printf "$(GREEN)  Go Web3 Wallet Backend - Makefile 帮助$(RESET)\n"
	@printf "$(GREEN)===========================================$(RESET)\n\n"
	@printf "$(YELLOW)可用的命令：$(RESET)\n\n"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## /  $(BLUE)/'  | sed 's/:/ $(RESET)- /'
	@printf "\n$(YELLOW)常用示例：$(RESET)\n"
	@printf "  $(BLUE)make dev$(RESET)         - 开发模式快速启动\n"
	@printf "  $(BLUE)make build$(RESET)       - 编译项目\n"
	@printf "  $(BLUE)make run$(RESET)         - 编译并运行\n"
	@printf "  $(BLUE)make test$(RESET)        - 运行测试\n"
	@printf "  $(BLUE)make fmt$(RESET)         - 格式化代码\n"
	@printf "\n"


## build: 编译应用程序 (当前平台)
build:
	@printf "$(BLUE)>>> 正在构建 $(APP_NAME) for $(GOOS)/$(GOARCH)...$(RESET)\n"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(TARGET) $(MAIN_PACKAGE)
	@printf "$(GREEN)✓ 构建成功: $(TARGET)$(RESET)\n"

## build-linux: 编译 Linux 版本
build-linux:
	@printf "$(BLUE)>>> 正在构建 Linux 版本...$(RESET)\n"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(MAIN_PACKAGE)
	@printf "$(GREEN)✓ Linux 构建完成: $(BUILD_DIR)/$(APP_NAME)-linux-amd64$(RESET)\n"

## build-darwin: 编译 macOS 版本
build-darwin:
	@printf "$(BLUE)>>> 正在构建 macOS 版本...$(RESET)\n"
	@mkdir -p $(BUILD_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	@printf "$(GREEN)✓ macOS 构建完成$(RESET)\n"


## run: 编译并运行 API 服务器
run: build
	@printf "$(BLUE)>>> 正在启动 API 服务器...$(RESET)\n"
	@$(TARGET) apiserver --config=./config.yaml

## dev: 直接运行 API 服务器（不编译），适合开发调试
dev:
	@printf "$(BLUE)>>> 正在以开发模式启动 API 服务器...$(RESET)\n"
	@go run $(MAIN_PACKAGE) apiserver --config=./config.yaml


## clean: 清理构建产物
clean:
	@printf "$(BLUE)>>> 正在清理构建产物...$(RESET)\n"
	@rm -rf $(BUILD_DIR)
	@rm -f $(COVERAGE_FILE)
	@go clean
	@printf "$(GREEN)✓ 清理完成$(RESET)\n"

## clean-all: 深度清理 (包括依赖缓存)
clean-all: clean
	@printf "$(BLUE)>>> 正在清理模块缓存...$(RESET)\n"
	@go clean -modcache
	@printf "$(GREEN)✓ 深度清理完成$(RESET)\n"


## test: 运行所有单元测试
test:
	@printf "$(BLUE)>>> 正在运行测试...$(RESET)\n"
	@go test -v -timeout $(TEST_TIMEOUT) ./...
	@printf "$(GREEN)✓ 测试完成$(RESET)\n"

## test-coverage: 运行测试并生成覆盖率报告
test-coverage:
	@printf "$(BLUE)>>> 正在运行测试并生成覆盖率...$(RESET)\n"
	@go test -v -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	@go tool cover -html=$(COVERAGE_FILE) -o coverage.html
	@printf "$(GREEN)✓ 测试覆盖率报告已生成: coverage.html$(RESET)\n"
	@go tool cover -func=$(COVERAGE_FILE) | grep total | awk '{print "总覆盖率: " $$3}'

## test-race: 运行竞态检测测试
test-race:
	@printf "$(BLUE)>>> 正在运行竞态检测测试...$(RESET)\n"
	@go test -race -timeout $(TEST_TIMEOUT) ./...
	@printf "$(GREEN)✓ 竞态检测完成$(RESET)\n"

## test-benchmark: 运行基准测试
test-benchmark:
	@printf "$(BLUE)>>> 正在运行基准测试...$(RESET)\n"
	@go test -bench=. -benchmem ./...
	@printf "$(GREEN)✓ 基准测试完成$(RESET)\n"

## lint: 运行代码静态检查
lint:
	@printf "$(BLUE)>>> 正在运行 golangci-lint...$(RESET)\n"
	@if command -v $(GOLANGCI_LINT) > /dev/null; then \
		$(GOLANGCI_LINT) run --timeout 5m ./...; \
		printf "$(GREEN)✓ Lint 检查完成$(RESET)\n"; \
	else \
		printf "$(RED)✗ 未找到 golangci-lint，请运行: make install-tools$(RESET)\n"; \
	fi

## lint-fix: 自动修复 lint 问题
lint-fix:
	@printf "$(BLUE)>>> 正在自动修复 lint 问题...$(RESET)\n"
	@if command -v $(GOLANGCI_LINT) > /dev/null; then \
		$(GOLANGCI_LINT) run --fix --timeout 5m ./...; \
		printf "$(GREEN)✓ Lint 自动修复完成$(RESET)\n"; \
	else \
		printf "$(RED)✗ 未找到 golangci-lint$(RESET)\n"; \
	fi

## format: 格式化代码
format: fmt

## fmt: 格式化所有 Go 代码
fmt:
	@printf "$(BLUE)>>> 正在格式化代码...$(RESET)\n"
	@printf "$(YELLOW)1/3 运行 gofmt...$(RESET)\n"
	@gofmt -s -w .
	@printf "$(YELLOW)2/3 运行 goimports-reviser...$(RESET)\n"
	@if command -v goimports-reviser > /dev/null; then \
		$(GOIMPORTS_REVISER_CMD) ./...; \
	else \
		printf "$(YELLOW)⚠ goimports-reviser 未安装，跳过$(RESET)\n"; \
	fi
	@printf "$(YELLOW)3/3 运行 golines...$(RESET)\n"
	@if command -v golines > /dev/null; then \
		$(GOLINES_CMD) $(SRC_DIRS); \
	else \
		printf "$(YELLOW)⚠ golines 未安装，跳过$(RESET)\n"; \
	fi
	@printf "$(GREEN)✓ 代码格式化完成$(RESET)\n"


## deps-tidy: 清理和整理依赖
deps-tidy:
	@printf "$(BLUE)>>> 正在整理依赖...$(RESET)\n"
	@go mod tidy
	@printf "$(GREEN)✓ 依赖整理完成$(RESET)\n"

## install-tools: 安装开发所需的工具
install-tools:
	@printf "$(BLUE)>>> 正在安装开发工具...$(RESET)\n"
	@printf "$(YELLOW)安装 golangci-lint...$(RESET)\n"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@printf "$(YELLOW)安装 golines...$(RESET)\n"
	@go install github.com/segmentio/golines@latest
	@printf "$(YELLOW)安装 goimports-reviser...$(RESET)\n"
	@go install github.com/incu6us/goimports-reviser/v3@latest