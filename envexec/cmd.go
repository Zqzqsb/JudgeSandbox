package envexec

import (
	"context"
	"fmt"
	"os"
)

// Runner 定义了运行命令的接口
type Runner interface {
	Start(context.Context) error
	Wait(context.Context) (*Status, error)
	Kill() error
}

/*
	命令构建器
	
	解耦配置和执行:
		Cmd 只包含配置信息
		Runner 负责实际的执行逻辑
		CmdBuilder 负责转换过程
	灵活性:
		可以有多种不同的执行环境（普通进程、沙箱、Docker等）
		每种环境都可以有自己的构建器实现
	资源管理：
		构建器可以在构建过程中分配必要的资源
		Runner 可以在执行完成后清理这些资
*/
type CmdBuilder interface {
	Build(*Cmd) (Runner, error)
}

// ExecveParam 定义了execve系统调用的参数
type ExecveParam struct {
	Args  []string         // 命令参数
	Env   []string         // 环境变量
	Files map[int]*os.File // 文件描述符映射
}

// CmdError 表示命令执行过程中发生的错误
type CmdError struct {
	Phase string // 错误发生的阶段
	Err   error  // 具体错误
}

func (e *CmdError) Error() string {
	return fmt.Sprintf("%s: %v", e.Phase, e.Err)
}

// ValidateCmd 验证命令配置是否合法
func ValidateCmd(cmd *Cmd) error {
	if len(cmd.Args) == 0 {
		return &CmdError{Phase: "validate", Err: fmt.Errorf("未指定命令")}
	}

	if cmd.TimeLimit < 0 {
		return &CmdError{Phase: "validate", Err: fmt.Errorf("时间限制不能为负")}
	}

	if cmd.MemoryLimit < 0 {
		return &CmdError{Phase: "validate", Err: fmt.Errorf("内存限制不能为负")}
	}

	if cmd.ProcLimit < 0 {
		return &CmdError{Phase: "validate", Err: fmt.Errorf("进程数限制不能为负")}
	}

	if cmd.CPURate < 0 || cmd.CPURate > 1 {
		return &CmdError{Phase: "validate", Err: fmt.Errorf("CPU使用率必须在0-1之间")}
	}

	return nil
}


/*
RunCmd 是评测系统的核心执行函数，它封装了完整的命令执行流程。

设计定位：
	作为高层次的封装函数，提供一站式的命令执行服务
	处理完整的执行生命周期：验证 → 准备 → 执行 → 收集
	统一的错误处理和资源管理
	支持上下文控制和取消操作
执行过程：
	验证阶段：检查命令参数的合法性
	准备阶段：将所需文件复制到执行环境中
	执行阶段：在隔离环境中运行命令
	收集阶段：收集执行结果和输出文件

*/

func RunCmd(ctx context.Context, env Environment, cmd *Cmd, copyIn CopyInReader, copyOut FileCollector) (*Status, error) {
	// 验证命令
	if err := ValidateCmd(cmd); err != nil {
		return nil, err
	}

	// 准备文件
	if err := PrepareFiles(ctx, cmd, copyIn); err != nil {
		return nil, &CmdError{Phase: "prepare", Err: err}
	}

	// 运行命令
	status, err := env.Run(ctx, cmd)
	if err != nil {
		return nil, &CmdError{Phase: "run", Err: err}
	}

	// 收集文件
	if len(cmd.CopyOut) > 0 {
		files, err := CollectFiles(ctx, cmd.CopyOut, copyOut)
		if err != nil {
			return nil, &CmdError{Phase: "collect", Err: err}
		}
		defer func() {
			for _, f := range files {
				if f != nil {
					f.Close()
				}
			}
		}()
	}
	return status, nil
}
