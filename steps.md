# build steps

## 安装插件

```bash
# 安装 protoc Go 插件
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
# 安装 gRPC Go 插件
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

## 生成 `protobuf` 代码

```bash
# 从 proto 文件生成 Go 代码，注意指定正确的模块路径
protoc --go_out=. --go_opt=module=github.com/zqzqsb/judgebox --go-grpc_out=. --go-grpc_opt=module=github.com/zqzqsb/judgebox pb/judge.proto
```