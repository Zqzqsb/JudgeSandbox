// Package envexec 提供了在受限环境（容器和cgroup）中运行程序的实用功能
package envexec

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// Status 定义了命令的执行状态
type Status struct {
	Status   ExitStatus    // 退出状态
	Error    string        // 错误信息
	Time     time.Duration // CPU时间
	Memory   int64         // 内存使用（字节）
	RunTime  time.Duration // 实际运行时间
	ExitCode int           // 退出码
}

// ExitStatus 定义了命令可能的退出状态
type ExitStatus int

const (
	// StatusNormal 表示命令正常退出
	StatusNormal ExitStatus = iota
	// StatusSignalled 表示命令被信号终止
	StatusSignalled
	// StatusREError 表示命令触发了运行时错误
	StatusREError
)

// FileError 定义了文件操作期间可能发生的错误
type FileError struct {
	Op   string // 操作类型
	Path string // 文件路径
	Err  error  // 具体错误
}

// 实现文件错误类的 error 接口
func (f *FileError) Error() string {
	return fmt.Sprintf("%s %s: %v", f.Op, f.Path, f.Err)
}

// Environment 定义了容器/沙箱环境的接口
type Environment interface {
	// Run 启动命令并等待完成
	Run(context.Context, *Cmd) (*Status, error)

	// CopyIn 将文件复制到容器内
	CopyIn(context.Context, *File) error

	// CopyOut 从容器中复制文件出来
	CopyOut(context.Context, *File) error

	// Close 释放资源
	io.Closer
}

// File 表示要复制进/出容器的文件
type File struct {
	Name string      // 文件名
	Path string      // 文件路径
	Mode os.FileMode // 文件权限
}

// Cmd 表示要在容器中执行的命令
type Cmd struct {
	Args        []string          // 命令参数
	Env         []string          // 环境变量
	Files       map[int]*os.File  // 文件描述符映射
	TTY         bool              // 是否分配TTY
	TimeLimit   time.Duration     // CPU时间限制
	MemoryLimit int64             // 内存限制（字节）
	ProcLimit   int               // 进程数限制
	CPURate     float64           // CPU使用率限制（0-1）
	CopyIn      map[string]string // 执行前要复制进容器的文件
	CopyOut     []string          // 执行后要复制出容器的文件
}

// NewCmd 创建一个带有默认值的新命令
func NewCmd(args []string) *Cmd {
	return &Cmd{
		Args:        args,
		Files:       make(map[int]*os.File),
		TimeLimit:   time.Second, // 默认时间限制1秒
		MemoryLimit: 256 << 20,   // 默认内存限制256MB
		ProcLimit:   1,           // 默认限制1个进程
		CPURate:     1.0,         // 默认CPU使用率100%
		CopyIn:      make(map[string]string),
	}
}
