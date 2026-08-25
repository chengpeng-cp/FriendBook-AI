# FriendBook AI 常用命令
.PHONY: build run test swag clean docker-up docker-down docker-logs

# 编译二进制到 bin/
build:
	go build -o bin/server cmd/server/main.go

# 本地运行
run:
	go run cmd/server/main.go

# 运行全部测试（详细信息 + 覆盖率）
test:
	go test ./... -v -cover

# 生成 Swagger 文档（需先 go install github.com/swaggo/swag/cmd/swag@latest）
swag:
	swag init -g cmd/server/main.go -o docs

# 清理构建产物
clean:
	rm -rf bin/ logs/

# Docker 一键部署
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f app
