## build steps

1. 环境准备与代码获取
   - 确保你安装了 Go 1.23 或更高版本，并配置好了 Go 开发环境。
   - 克隆或下载项目源码（如果尚未获取代码，请从官方仓库或源码包获取）。

2. 依赖安装与环境配置
   - 进入项目根目录，运行 `go mod tidy` 来安装所有依赖。
   - 检查 `go.mod` 文件，确保所有依赖版本和包都符合要求。

3. 定义并生成 Protobuf 代码
   - 阅读 `pb/judge.proto` 文件，了解服务接口定义（例如 `Executor` 服务、`Request` 和 `Response` 消息）。
   - 使用 protoc 编译 .proto 文件。示例命令：
     ```bash
     protoc --go_out=. --go-grpc_out=. pb/judge.proto
     ```
   - 确认生成的 `judge.pb.go` 和 `judge_grpc.pb.go` 文件生成无误。

4. 阅读项目结构与核心模块
   - 浏览 `cmd/` 目录，理解各个入口命令及其作用。
   - 检查 `envexec/` 和 `worker/` 目录，理解程序执行、资源限制和任务调度的实现逻辑。
   - 阅读 `README.md` 和 `README.cn.md` 获取部署、运行和 REST API 使用说明。

5. 配置与业务逻辑完善
   - 检查项目中的配置文件，如 `mount.yaml`、`.air.toml`、Dockerfile 系列文件，理解系统启动及容器相关配置。
   - 根据需要修改配置，确保本地环境与项目要求相符。

6. 编写并运行单元测试
   - 查看 `test/` 目录及相关测试代码，对各个模块的功能进行单元测试。
   - 运行测试命令：
     ```bash
     go test ./...
     ```
   - 调试测试失败项，确保所有模块功能正确。

7. 构建项目
   - 构建项目可执行文件：
     ```bash
     go build -o go-judge .
     ```
   - 如果需要 Docker 部署，可参考 `Dockerfile.alpine` 或其他 Dockerfile 文件构建镜像，例如：
     ```bash
     docker build -f Dockerfile.alpine -t go-judge-alpine .
     ```

8. 部署与验证
   - 运行生成的可执行文件或 Docker 容器，根据 `README` 中描述的 REST API 接口（如 `/run`, `/file`, `/ws`, `/stream` 等）进行测试。
   - 检查日志，验证服务运行情况与资源限制、文件操作等功能是否正常。

按照上述步骤，你将能够完整地复现原始的 go-judge 项目，确保各项模块和接口按照源码实现工作。
