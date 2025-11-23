package main

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestPipe(t *testing.T) {
	// 创建管道：pipe[0]是读端，pipe[1]是写端
	pipeReader, pipeWriter, err := os.Pipe()
	if err != nil {
		log.Fatalf("创建管道失败: %v", err)
	}
	defer pipeReader.Close()
	defer pipeWriter.Close()

	// 第一个命令：例如 "ls -l"（列出当前目录文件）
	cmd1 := exec.Command("ls", "-l")
	// 将cmd1的标准输出重定向到管道的写端
	cmd1.Stdout = pipeWriter
	// 可选：将cmd1的标准错误输出到控制台
	cmd1.Stderr = os.Stderr

	// 第二个命令：例如 "grep go"（过滤包含"go"的行）
	cmd2 := exec.Command("grep", "go")
	// 将cmd2的标准输入重定向到管道的读端
	cmd2.Stdin = pipeReader
	// 将cmd2的标准输出和错误输出到控制台
	cmd2.Stdout = os.Stdout
	cmd2.Stderr = os.Stderr

	// 启动第一个命令
	if err := cmd1.Start(); err != nil {
		log.Fatalf("启动cmd1失败: %v", err)
	}

	// 启动第二个命令
	if err := cmd2.Start(); err != nil {
		log.Fatalf("启动cmd2失败: %v", err)
	}

	// 等待第一个命令执行完成，然后关闭管道写端（否则cmd2会一直等待输入）
	if err := cmd1.Wait(); err != nil {
		log.Printf("cmd1执行失败: %v", err)
	}
	pipeWriter.Close() // 必须关闭写端，否则cmd2会阻塞

	// 等待第二个命令执行完成
	if err := cmd2.Wait(); err != nil {
		log.Printf("cmd2执行失败: %v", err)
	}
}
