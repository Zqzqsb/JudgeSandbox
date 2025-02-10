package envexec

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// MockEnvironment 模拟执行环境
type MockEnvironment struct {
	RunFunc    func(context.Context, *Cmd) (*Status, error)
	KillCalled bool
}

func (m *MockEnvironment) Run(ctx context.Context, cmd *Cmd) (*Status, error) {
	return m.RunFunc(ctx, cmd)
}

func (m *MockEnvironment) CopyIn(ctx context.Context, file *File) error {
	return nil
}

func (m *MockEnvironment) CopyOut(ctx context.Context, file *File) error {
	return nil
}

func (m *MockEnvironment) Close() error {
	return nil
}

// MockFileReader 模拟文件读取器
type MockFileReader struct {
	OpenFileFunc func(context.Context, string) (*os.File, error)
}

func (m *MockFileReader) OpenFile(ctx context.Context, path string) (*os.File, error) {
	return m.OpenFileFunc(ctx, path)
}

// MockFileCollector 模拟文件收集器
type MockFileCollector struct {
	CollectFileFunc  func(context.Context, string) (*os.File, error)
	CollectFilesFunc func(context.Context, []string) (map[string]*os.File, error)
}

func (m *MockFileCollector) CollectFile(ctx context.Context, path string) (*os.File, error) {
	return m.CollectFileFunc(ctx, path)
}

func (m *MockFileCollector) CollectFiles(ctx context.Context, paths []string) (map[string]*os.File, error) {
	return m.CollectFilesFunc(ctx, paths)
}

func TestRunCmd(t *testing.T) {
	// 创建临时文件用于测试
	tmpFile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	tests := []struct {
		name    string
		cmd     *Cmd
		wantErr bool
		setup   func() (*MockEnvironment, *MockFileReader, *MockFileCollector)
	}{
		{
			name: "正常执行",
			cmd: &Cmd{
				Args:        []string{"/bin/echo", "hello"},
				TimeLimit:   time.Second,
				MemoryLimit: 64 << 20,
				CopyIn: map[string]string{
					"input.txt": tmpFile.Name(),
				},
				CopyOut: []string{"output.txt"},
			},
			wantErr: false,
			setup: func() (*MockEnvironment, *MockFileReader, *MockFileCollector) {
				env := &MockEnvironment{
					RunFunc: func(ctx context.Context, cmd *Cmd) (*Status, error) {
						return &Status{
							Time:     100 * time.Millisecond,
							Memory:   1024,
							RunTime:  150 * time.Millisecond,
							ExitCode: 0,
						}, nil
					},
				}
				reader := &MockFileReader{
					OpenFileFunc: func(ctx context.Context, path string) (*os.File, error) {
						return tmpFile, nil
					},
				}
				collector := &MockFileCollector{
					CollectFilesFunc: func(ctx context.Context, paths []string) (map[string]*os.File, error) {
						return map[string]*os.File{"output.txt": tmpFile}, nil
					},
				}
				return env, reader, collector
			},
		},
		{
			name: "命令验证失败",
			cmd: &Cmd{
				Args:        []string{},
				TimeLimit:   -1,
				MemoryLimit: -1,
			},
			wantErr: true,
			setup: func() (*MockEnvironment, *MockFileReader, *MockFileCollector) {
				return &MockEnvironment{}, &MockFileReader{}, &MockFileCollector{}
			},
		},
		{
			name: "文件准备失败",
			cmd: &Cmd{
				Args:        []string{"/bin/echo", "hello"},
				TimeLimit:   time.Second,
				MemoryLimit: 64 << 20,
				CopyIn: map[string]string{
					"input.txt": "不存在的文件.txt",
				},
			},
			wantErr: true,
			setup: func() (*MockEnvironment, *MockFileReader, *MockFileCollector) {
				reader := &MockFileReader{
					OpenFileFunc: func(ctx context.Context, path string) (*os.File, error) {
						return nil, fmt.Errorf("文件不存在")
					},
				}
				return &MockEnvironment{}, reader, &MockFileCollector{}
			},
		},
		{
			name: "执行超时",
			cmd: &Cmd{
				Args:        []string{"/bin/sleep", "10"},
				TimeLimit:   time.Second,
				MemoryLimit: 64 << 20,
			},
			wantErr: true,
			setup: func() (*MockEnvironment, *MockFileReader, *MockFileCollector) {
				env := &MockEnvironment{
					RunFunc: func(ctx context.Context, cmd *Cmd) (*Status, error) {
						return nil, context.DeadlineExceeded
					},
				}
				return env, &MockFileReader{}, &MockFileCollector{}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			env, reader, collector := tt.setup()

			_, err := RunCmd(ctx, env, tt.cmd, reader, collector)
			if (err != nil) != tt.wantErr {
				t.Errorf("RunCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateCmd(t *testing.T) {
	tests := []struct {
		name    string
		cmd     *Cmd
		wantErr bool
	}{
		{
			name: "有效命令",
			cmd: &Cmd{
				Args:        []string{"/bin/echo", "hello"},
				TimeLimit:   time.Second,
				MemoryLimit: 64 << 20,
				ProcLimit:   1,
				CPURate:     1.0,
			},
			wantErr: false,
		},
		{
			name: "空命令",
			cmd: &Cmd{
				Args:        []string{},
				TimeLimit:   time.Second,
				MemoryLimit: 64 << 20,
			},
			wantErr: true,
		},
		{
			name: "负时间限制",
			cmd: &Cmd{
				Args:        []string{"/bin/echo"},
				TimeLimit:   -1,
				MemoryLimit: 64 << 20,
			},
			wantErr: true,
		},
		{
			name: "负内存限制",
			cmd: &Cmd{
				Args:        []string{"/bin/echo"},
				TimeLimit:   time.Second,
				MemoryLimit: -1,
			},
			wantErr: true,
		},
		{
			name: "无效CPU使用率",
			cmd: &Cmd{
				Args:        []string{"/bin/echo"},
				TimeLimit:   time.Second,
				MemoryLimit: 64 << 20,
				CPURate:     2.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCmd(tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewCmd(t *testing.T) {
	args := []string{"/bin/echo", "hello"}
	cmd := NewCmd(args)

	if len(cmd.Args) != len(args) {
		t.Errorf("NewCmd().Args = %v, want %v", cmd.Args, args)
	}

	if cmd.TimeLimit != time.Second {
		t.Errorf("NewCmd().TimeLimit = %v, want %v", cmd.TimeLimit, time.Second)
	}

	if cmd.MemoryLimit != 256<<20 {
		t.Errorf("NewCmd().MemoryLimit = %v, want %v", cmd.MemoryLimit, 256<<20)
	}

	if cmd.ProcLimit != 1 {
		t.Errorf("NewCmd().ProcLimit = %v, want %v", cmd.ProcLimit, 1)
	}

	if cmd.CPURate != 1.0 {
		t.Errorf("NewCmd().CPURate = %v, want %v", cmd.CPURate, 1.0)
	}
}
