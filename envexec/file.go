package envexec

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

// FileCollector 定义了收集文件的接口
type FileCollector interface {
	// CollectFile 收集单个文件
	CollectFile(context.Context, string) (*os.File, error)

	// CollectFiles 收集多个文件
	CollectFiles(context.Context, []string) (map[string]*os.File, error)
}

// FileWriter 定义了写入文件的接口
type FileWriter interface {
	// WriteFile 写入单个文件
	WriteFile(context.Context, string, *os.File) error
}

// CopyInReader 定义了用于复制进容器的文件读取接口
type CopyInReader interface {
	// OpenFile 打开文件用于读取
	OpenFile(context.Context, string) (*os.File, error)
}

// CopyOutWriter 定义了用于复制出容器的文件写入接口
type CopyOutWriter interface {
	// CreateFile 创建文件用于写入
	CreateFile(context.Context, string, os.FileMode) (*os.File, error)
}


// --------------------- 工具函数 --------------------- 

// CopyFile 将文件从src复制到dst
func CopyFile(dst *os.File, src *os.File) error {
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}
	return dst.Sync()
}

// EnsureFilePath 确保文件的父目录存在
func EnsureFilePath(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}

// CloseFiles 关闭多个文件
func CloseFiles(files map[int]*os.File) {
	for _, f := range files {
		if f != nil {
			f.Close()
		}
	}
}

// PrepareFiles 为命令执行准备文件
// 参数:
//   - ctx: 上下文，用于控制操作的超时和取消
//   - cmd: 要执行的命令，包含了需要复制的文件信息
//   - copyIn: 文件读取接口，用于从外部读取源文件
// 返回:
//   - error: 如果发生错误，返回被 FileError 包装的具体错误
func PrepareFiles(ctx context.Context, cmd *Cmd, copyIn CopyInReader) error {
    // 遍历所有需要复制的文件
    // dst: 目标文件在容器中的路径
    // src: 源文件在外部的路径
    for dst, src := range cmd.CopyIn {
        // 步骤1: 打开源文件
        // 使用提供的 copyIn 接口打开源文件，这样可以支持不同的文件源（本地文件、远程文件等）
        srcFile, err := copyIn.OpenFile(ctx, src)
        if err != nil {
            return &FileError{Op: "open", Path: src, Err: err}
        }
        // 确保在函数返回时关闭源文件，防止文件句柄泄漏
        defer srcFile.Close()

        // 步骤2: 确保目标文件的父目录存在
        // 如果目标路径的父目录不存在，会创建所需的目录
        if err := EnsureFilePath(dst); err != nil {
            return &FileError{Op: "mkdir", Path: dst, Err: err}
        }

        // 步骤3: 创建目标文件
        // os.O_WRONLY: 只写模式
        // os.O_CREATE: 如果文件不存在则创建
        // os.O_TRUNC: 如果文件存在则清空内容
        // 0644: 文件权限 (rw-r--r--)
        dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
        if err != nil {
            return &FileError{Op: "create", Path: dst, Err: err}
        }
        // 确保在函数返回时关闭目标文件
        defer dstFile.Close()

        // 步骤4: 复制文件内容
        // 将源文件的内容完整地复制到目标文件
        if err := CopyFile(dstFile, srcFile); err != nil {
            return &FileError{Op: "copy", Path: dst, Err: err}
        }
    }

    return nil
}

// CollectFiles 在命令执行后收集文件
func CollectFiles(ctx context.Context, copyOut []string, collector FileCollector) (map[string]*os.File, error) {
	if len(copyOut) == 0 {
		return nil, nil
	}
	return collector.CollectFiles(ctx, copyOut)
}
