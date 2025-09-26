GIT_COMMIT ?= $(shell git rev-parse --short HEAD)

a:
	esbuild ./browser/index.ts --bundle --minify --sourcemap --outfile=main/assets/zhteuern.js

w:
	@go run watch/fw/fw.go

b:
	@go build -ldflags="-w -s -X main.environment=prod -X main.gitCommit=${GIT_COMMIT}" -o bin/main ./main/*.go

run:
	@go run -ldflags="-w -s -X main.environment=dev -X main.gitCommit=${GIT_COMMIT} -X main.host=localhost:3000" ./main/*.go

# Compress binary
u:
	upx bin/main