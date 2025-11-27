# ==================================================================================== #
# Makefile for Go Web3 Wallet Backend
# ==================================================================================== #

# ==================================================================================== #
# 全局变量配置
# ==================================================================================== #

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

# ==================================================================================== #
# PHONY 目标声明
# ==================================================================================== #

.PHONY: all help
.PHONY: build build-linux build-darwin
.PHONY: run dev
.PHONY: clean clean-all
.PHONY: test test-coverage test-race test-benchmark
.PHONY: lint lint-fix format fmt
.PHONY: deps deps-download deps-verify deps-tidy deps-upgrade
.PHONY: install-tools
.PHONY: docker-build docker-run docker-stop
.PHONY: db-migrate db-rollback db-status
.PHONY: version info

# ==================================================================================== #
# 默认目标
# ==================================================================================== #

all: help

# ==================================================================================== #
# 构建命令
# ==================================================================================== #

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


# ==================================================================================== #
# 运行命令
# ==================================================================================== #

## run: 编译并运行应用 (生产模式)
run: build
	@printf "$(BLUE)>>> 正在启动服务器...$(RESET)\n"
	@$(TARGET) serve

## dev: 使用热重载运行开发服务器 (需要安装 air)
dev:
	@printf "$(BLUE)>>> 正在启动开发服务器 (热重载)...$(RESET)\n"
	@if command -v air > /dev/null; then \
		air; \
	else \
		printf "$(RED)✗ 未找到 air 工具，请运行: go install github.com/air-verse/air@latest$(RESET)\n"; \
		printf "$(YELLOW)>>> 降级为普通运行模式...$(RESET)\n"; \
		$(MAKE) run; \
	fi

# ==================================================================================== #
# 清理命令
# ==================================================================================== #

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

# ==================================================================================== #
# 测试命令
# ==================================================================================== #

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

# ==================================================================================== #
# 代码质量命令
# ==================================================================================== #

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

# ==================================================================================== #
# 依赖管理命令
# ==================================================================================== #

## deps: 安装项目依赖
deps: deps-download

## deps-download: 下载依赖包
deps-download:
	@printf "$(BLUE)>>> 正在下载依赖...$(RESET)\n"
	@go mod download
	@printf "$(GREEN)✓ 依赖下载完成$(RESET)\n"

## deps-verify: 验证依赖完整性
deps-verify:
	@printf "$(BLUE)>>> 正在验证依赖...$(RESET)\n"
	@go mod verify
	@printf "$(GREEN)✓ 依赖验证通过$(RESET)\n"

## deps-tidy: 清理和整理依赖
deps-tidy:
	@printf "$(BLUE)>>> 正在整理依赖...$(RESET)\n"
	@go mod tidy
	@printf "$(GREEN)✓ 依赖整理完成$(RESET)\n"

## deps-upgrade: 升级所有依赖到最新版本
deps-upgrade:
	@printf "$(BLUE)>>> 正在升级依赖...$(RESET)\n"
	@go get -u ./...
	@go mod tidy
	@printf "$(GREEN)✓ 依赖升级完成$(RESET)\n"

# ==================================================================================== #
# 工具安装
# ==================================================================================== #

## install-tools: 安装开发所需的工具
install-tools:
	@printf "$(BLUE)>>> 正在安装开发工具...$(RESET)\n"
	@printf "$(YELLOW)安装 golangci-lint...$(RESET)\n"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@printf "$(YELLOW)安装 golines...$(RESET)\n"
	@go install github.com/segmentio/golines@latest
	@printf "$(YELLOW)安装 goimports-reviser...$(RESET)\n"
	@go install github.com/incu6us/goimports-reviser/v3@latest
	@printf "$(YELLOW)安装 air (热重载)...$(RESET)\n"
	@go install github.com/air-verse/air@latest
	@printf "$(YELLOW)安装 mockgen (Mock 生成)...$(RESET)\n"
	@go install go.uber.org/mock/mockgen@latest
	@printf "$(GREEN)✓ 工具安装完成$(RESET)\n"

# ==================================================================================== #
# Docker 命令
# ==================================================================================== #

## docker-build: 构建 Docker 镜像
docker-build:
	@printf "$(BLUE)>>> 正在构建 Docker 镜像...$(RESET)\n"
	@docker build -t $(APP_NAME):$(VERSION) -t $(APP_NAME):latest .
	@printf "$(GREEN)✓ Docker 镜像构建完成$(RESET)\n"

## docker-run: 运行 Docker 容器
docker-run:
	@printf "$(BLUE)>>> 正在启动 Docker 容器...$(RESET)\n"
	@docker run -d -p 8080:8080 --name $(APP_NAME) $(APP_NAME):latest
	@printf "$(GREEN)✓ Docker 容器已启动$(RESET)\n"

## docker-stop: 停止并删除 Docker 容器
docker-stop:
	@printf "$(BLUE)>>> 正在停止 Docker 容器...$(RESET)\n"
	@docker stop $(APP_NAME) 2>/dev/null || true
	@docker rm $(APP_NAME) 2>/dev/null || true
	@printf "$(GREEN)✓ Docker 容器已停止$(RESET)\n"

# ==================================================================================== #
# 信息命令
# ==================================================================================== #

## version: 显示版本信息
version:
	@printf "$(GREEN)应用名称:$(RESET)    $(APP_NAME)\n"
	@printf "$(GREEN)版本:$(RESET)        $(VERSION)\n"
	@printf "$(GREEN)Git Commit:$(RESET)  $(GIT_COMMIT)\n"
	@printf "$(GREEN)构建时间:$(RESET)    $(BUILD_TIME)\n"
	@printf "$(GREEN)Go 版本:$(RESET)     $(shell go version)\n"

## info: 显示项目信息
info: version
	@printf "$(GREEN)项目模块:$(RESET)    $(PROJECT_MODULE)\n"
	@printf "$(GREEN)目标平台:$(RESET)    $(GOOS)/$(GOARCH)\n"
	@printf "$(GREEN)构建目录:$(RESET)    $(BUILD_DIR)\n"

## help: 显示帮助信息
help:
	@printf "$(BLUE)╔════════════════════════════════════════════════════════════════╗$(RESET)\n"
	@printf "$(BLUE)║         Go Web3 Wallet Backend - Makefile 帮助                ║$(RESET)\n"
	@printf "$(BLUE)╚════════════════════════════════════════════════════════════════╝$(RESET)\n"
	@printf "\n"
	@printf "$(GREEN)使用方式:$(RESET)\n"
	@printf "  make $(YELLOW)<target>$(RESET)\n"
	@printf "\n"
	@printf "$(GREEN)🔨 构建命令:$(RESET)\n"
	@printf "  $(YELLOW)build$(RESET)              编译应用程序 (当前平台)\n"
	@printf "  $(YELLOW)build-linux$(RESET)        编译 Linux 版本\n"
	@printf "  $(YELLOW)build-darwin$(RESET)       编译 macOS 版本\n"
	@printf "  $(YELLOW)build-windows$(RESET)      编译 Windows 版本\n"
	@printf "  $(YELLOW)build-all$(RESET)          编译所有平台版本\n"
	@printf "\n"
	@printf "$(GREEN)🚀 运行命令:$(RESET)\n"
	@printf "  $(YELLOW)run$(RESET)                编译并运行应用\n"
	@printf "  $(YELLOW)dev$(RESET)                使用热重载运行开发服务器\n"
	@printf "\n"
	@printf "$(GREEN)🧪 测试命令:$(RESET)\n"
	@printf "  $(YELLOW)test$(RESET)               运行所有单元测试\n"
	@printf "  $(YELLOW)test-coverage$(RESET)      生成测试覆盖率报告\n"
	@printf "  $(YELLOW)test-race$(RESET)          运行竞态检测测试\n"
	@printf "  $(YELLOW)test-benchmark$(RESET)     运行基准测试\n"
	@printf "\n"
	@printf "$(GREEN)✨ 代码质量:$(RESET)\n"
	@printf "  $(YELLOW)lint$(RESET)               运行代码静态检查\n"
	@printf "  $(YELLOW)lint-fix$(RESET)           自动修复 lint 问题\n"
	@printf "  $(YELLOW)fmt$(RESET)                格式化所有代码\n"
	@printf "\n"
	@printf "$(GREEN)📦 依赖管理:$(RESET)\n"
	@printf "  $(YELLOW)deps$(RESET)               下载项目依赖\n"
	@printf "  $(YELLOW)deps-tidy$(RESET)          清理和整理依赖\n"
	@printf "  $(YELLOW)deps-verify$(RESET)        验证依赖完整性\n"
	@printf "  $(YELLOW)deps-upgrade$(RESET)       升级所有依赖\n"
	@printf "\n"
	@printf "$(GREEN)🛠️  工具安装:$(RESET)\n"
	@printf "  $(YELLOW)install-tools$(RESET)      安装开发所需工具\n"
	@printf "\n"
	@printf "$(GREEN)🐳 Docker 命令:$(RESET)\n"
	@printf "  $(YELLOW)docker-build$(RESET)       构建 Docker 镜像\n"
	@printf "  $(YELLOW)docker-run$(RESET)         运行 Docker 容器\n"
	@printf "  $(YELLOW)docker-stop$(RESET)        停止并删除容器\n"
	@printf "\n"
	@printf "$(GREEN)🧹 清理命令:$(RESET)\n"
	@printf "  $(YELLOW)clean$(RESET)              清理构建产物\n"
	@printf "  $(YELLOW)clean-all$(RESET)          深度清理 (包括缓存)\n"
	@printf "\n"
	@printf "$(GREEN)ℹ️  信息命令:$(RESET)\n"
	@printf "  $(YELLOW)version$(RESET)            显示版本信息\n"
	@printf "  $(YELLOW)info$(RESET)               显示项目信息\n"
	@printf "  $(YELLOW)help$(RESET)               显示此帮助信息\n"
	@printf "\n"
	@printf "$(BLUE)更多信息请访问: $(PROJECT_MODULE)$(RESET)\n"
	@printf "\n"