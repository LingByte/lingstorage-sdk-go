# LingStorage SDK Makefile

# 变量定义
MODULE_NAME=github.com/LingByte/lingstorage-sdk
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# 默认目标
.DEFAULT_GOAL := help

# 帮助信息
.PHONY: help
help: ## 显示帮助信息
	@echo "LingStorage SDK 开发工具"
	@echo ""
	@echo "可用命令:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# 初始化项目
.PHONY: init
init: ## 初始化项目依赖
	@echo "📦 初始化项目依赖..."
	@go mod tidy
	@go mod download
	@echo "✅ 依赖初始化完成"

# 运行测试
.PHONY: test
test: ## 运行所有测试
	@echo "🧪 运行测试..."
	@go test -v ./...

# 运行测试并生成覆盖率报告
.PHONY: test-coverage
test-coverage: ## 运行测试并生成覆盖率报告
	@echo "📊 运行测试并生成覆盖率报告..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ 覆盖率报告已生成: coverage.html"

# 运行基准测试
.PHONY: bench
bench: ## 运行基准测试
	@echo "⚡ 运行基准测试..."
	@go test -bench=. -benchmem ./...

# 代码检查
.PHONY: lint
lint: ## 代码检查
	@echo "🔍 代码检查..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过代码检查"; \
		echo "💡 安装命令: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 格式化代码
.PHONY: fmt
fmt: ## 格式化代码
	@echo "✨ 格式化代码..."
	@go fmt ./...
	@goimports -w . 2>/dev/null || echo "💡 建议安装 goimports: go install golang.org/x/tools/cmd/goimports@latest"

# 清理
.PHONY: clean
clean: ## Clean build files
	@echo "🧹 Cleaning build files..."
	@rm -f coverage.out coverage.html
	@go clean -cache
	@echo "✅ Cleaning completed"

# 运行演示
.PHONY: demo
demo: ## 运行 SDK 演示
	@echo "🚀 运行 SDK 演示..."
	@if [ -z "$$LINGSTORAGE_API_KEY" ] || [ -z "$$LINGSTORAGE_API_SECRET" ]; then \
		echo "⚠️  请设置环境变量:"; \
		echo "  export LINGSTORAGE_API_KEY=your-api-key"; \
		echo "  export LINGSTORAGE_API_SECRET=your-api-secret"; \
		echo "  export LINGSTORAGE_BASE_URL=http://localhost:7075  # 可选"; \
		exit 1; \
	fi
	@cd examples/demo && go run demo.go

# 运行服务器演示
.PHONY: demo-server
demo-server: ## 运行服务器端演示（连接到 localhost:7075）
	@echo "🚀 运行服务器端演示..."
	@cd server && go run demo.go

# 运行示例
.PHONY: example-basic
example-basic: ## Run basic upload example
	@echo "🚀 Running basic upload example..."
	@cd examples/basic_upload && go run main.go

.PHONY: example-batch
example-batch: ## 运行批量上传示例
	@echo "🚀 运行批量上传示例..."
	@cd examples/batch_upload && go run main.go

.PHONY: example-image
example-image: ## 运行图片处理示例
	@echo "🚀 运行图片处理示例..."
	@cd examples/image_processing && go run main.go

.PHONY: example-progress
example-progress: ## 运行进度监控示例
	@echo "🚀 运行进度监控示例..."
	@cd examples/progress_monitoring && go run main.go

# 构建示例
.PHONY: build-examples
build-examples: ## Build all examples
	@echo "🔨 Building example programs..."
	@mkdir -p bin
	@cd examples/basic_upload && go build -o ../../bin/basic_upload main.go
	@cd examples/batch_upload && go build -o ../../bin/batch_upload main.go
	@cd examples/image_processing && go build -o ../../bin/image_processing main.go
	@cd examples/progress_monitoring && go build -o ../../bin/progress_monitoring main.go
	@echo "✅ Example programs built, located in bin/ directory"

# 安装开发工具
.PHONY: install-tools
install-tools: ## 安装开发工具
	@echo "🛠️  安装开发工具..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/stretchr/testify@latest
	@echo "✅ 开发工具安装完成"

# 检查依赖
.PHONY: check-deps
check-deps: ## 检查依赖
	@echo "🔍 检查依赖..."
	@go version
	@go mod verify
	@echo "✅ 依赖检查完成"

# 更新依赖
.PHONY: update-deps
update-deps: ## 更新依赖
	@echo "📦 更新依赖..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ 依赖更新完成"

# 生成文档
.PHONY: docs
docs: ## 生成文档
	@echo "📚 生成文档..."
	@if command -v godoc >/dev/null 2>&1; then \
		echo "启动文档服务器: http://localhost:6060/pkg/$(MODULE_NAME)"; \
		godoc -http=:6060; \
	else \
		echo "⚠️  godoc 未安装"; \
		echo "💡 安装命令: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# 发布检查
.PHONY: release-check
release-check: clean fmt lint test ## Pre-release check
	@echo "🚀 Pre-release check..."
	@echo "✅ All checks passed, ready to release"

# 创建测试文件
.PHONY: create-test-files
create-test-files: ## Create test files
	@echo "📁 Creating test files..."
	@mkdir -p testdata
	@echo "Hello, World!" > testdata/test.txt
	@echo "This is a test file for SDK." > testdata/readme.txt
	@echo "Binary data test" > testdata/binary.dat
	@echo "✅ Test files created, located in testdata/ directory"

# 运行集成测试
.PHONY: integration-test
integration-test: create-test-files ## 运行集成测试
	@echo "🔗 运行集成测试..."
	@if [ -z "$$LINGSTORAGE_BASE_URL" ] || [ -z "$$LINGSTORAGE_API_KEY" ]; then \
		echo "⚠️  请设置环境变量:"; \
		echo "  export LINGSTORAGE_BASE_URL=http://your-server:7075"; \
		echo "  export LINGSTORAGE_API_KEY=your-api-key"; \
		exit 1; \
	fi
	@echo "测试服务器: $$LINGSTORAGE_BASE_URL"
	@cd examples/basic_upload && go run main.go ../../testdata/test.txt
	@echo "✅ 集成测试完成"

# 版本信息
.PHONY: version
version: ## Display version information
	@echo "Version: $(VERSION)"
	@echo "Module: $(MODULE_NAME)"
	@go version