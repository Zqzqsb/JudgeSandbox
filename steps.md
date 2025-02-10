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

## 实现 envexec 包

1. 创建 `interface.go`
   - 定义了包的核心接口和类型
   - 包括 Environment 接口，用于容器/沙箱环境
   - 定义了 Status、File、Cmd 等基本类型
   - 实现了基本的错误处理机制

2. 创建 `file.go`
   - 实现文件操作相关的接口（FileCollector、FileWriter、CopyInReader、CopyOutWriter）
   - 提供文件复制、路径确保等基础功能
   - 实现文件准备（PrepareFiles）和收集（CollectFiles）功能

3. 创建 `cmd.go`
   - 实现命令执行相关的接口（Runner、CmdBuilder）
   - 提供命令验证和执行的核心功能
   - 实现默认命令配置和错误处理机制
   - 集成文件操作和命令执行流程
