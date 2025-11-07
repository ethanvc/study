package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/net"
)

func main() {
	// 检查命令行参数
	if len(os.Args) < 2 {
		fmt.Println("用法: tcpstate <远端端口1> <远端端口2> ...")
		fmt.Println("示例: tcpstate 80 443 8080")
		os.Exit(1)
	}

	// 获取要监控的远端端口列表
	remotePorts := os.Args[1:]

	// 验证端口
	for _, portStr := range remotePorts {
		port, err := strconv.Atoi(portStr)
		if err != nil || port < 1 || port > 65535 {
			fmt.Fprintf(os.Stderr, "错误: 无效的端口号 '%s' (端口范围: 1-65535)\n", portStr)
			os.Exit(1)
		}
	}

	fmt.Printf("开始监控远端端口: %s\n", strings.Join(remotePorts, ", "))
	fmt.Println("每2秒更新一次统计数据")
	fmt.Println("按 Ctrl+C 退出监控")

	// 主循环
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	stats := collectTCPStats(remotePorts)
	printTable(stats)

	// 定时执行
	for range ticker.C {
		stats := collectTCPStats(remotePorts)
		printTable(stats)
	}
}

// TCP状态常量
var tcpStates = []string{
	"ESTABLISHED",
	"SYN_SENT",
	"SYN_RECV",
	"FIN_WAIT1",
	"FIN_WAIT2",
	"TIME_WAIT",
	"CLOSE",
	"CLOSE_WAIT",
	"LAST_ACK",
	"LISTEN",
	"CLOSING",
}

// 统计结果
type TCPStats struct {
	Port   uint32
	States map[string]int
}

// 收集TCP连接统计信息
func collectTCPStats(remotePortsStr []string) []TCPStats {
	// 将端口字符串转换为整数集合
	remotePorts := make(map[uint32]bool)
	for _, portStr := range remotePortsStr {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}
		remotePorts[uint32(port)] = true
	}

	// 获取所有TCP连接
	connections, err := net.Connections("tcp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "获取TCP连接失败: %v\n", err)
		return nil
	}

	// 统计数据结构：port -> state -> count
	statsMap := make(map[uint32]map[string]int)

	for _, conn := range connections {
		port := conn.Raddr.Port
		if port == 0 {
			// for listen state, only have laddr.port
			port = conn.Laddr.Port
		}
		if !remotePorts[port] {
			continue
		}

		// 初始化端口统计
		if statsMap[port] == nil {
			statsMap[port] = make(map[string]int)
		}

		// 统计状态
		state := conn.Status
		statsMap[port][state]++
	}

	// 转换为切片结果
	var results []TCPStats
	for _, portStr := range remotePortsStr {
		port, _ := strconv.Atoi(portStr)
		portUint := uint32(port)

		stats := TCPStats{
			Port:   portUint,
			States: make(map[string]int),
		}

		if portStats, exists := statsMap[portUint]; exists {
			stats.States = portStats
		}

		results = append(results, stats)
	}

	return results
}

// 打印表格
func printTable(stats []TCPStats) {
	// 打印时间戳
	fmt.Printf("\n═══════════════════════════════════════════════════════════════\n")
	fmt.Printf("TCP状态统计报告 - %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")

	if len(stats) == 0 {
		fmt.Println("没有统计数据")
		return
	}

	// 打印表头
	fmt.Printf("%-10s", "端口")
	for _, state := range tcpStates {
		fmt.Printf(" | %-12s", state)
	}
	fmt.Println()

	// 打印分隔线
	fmt.Print(strings.Repeat("-", 10))
	for range tcpStates {
		fmt.Print(strings.Repeat("-", 16))
	}
	fmt.Println()

	// 打印每个端口的统计数据
	totalByState := make(map[string]int)
	totalConnections := 0

	for _, portStats := range stats {
		// 计算该端口的总连接数
		portTotal := 0
		for _, count := range portStats.States {
			portTotal += count
		}

		// 如果该端口没有连接，跳过
		if portTotal == 0 {
			continue
		}

		fmt.Printf("%-10d", portStats.Port)
		for _, state := range tcpStates {
			count := portStats.States[state]
			if count > 0 {
				fmt.Printf(" | %-12d", count)
				totalByState[state] += count
				totalConnections += count
			} else {
				fmt.Printf(" | %-12s", "-")
			}
		}
		fmt.Println()
	}

	// 打印总计行
	if totalConnections > 0 {
		fmt.Print(strings.Repeat("-", 10))
		for range tcpStates {
			fmt.Print(strings.Repeat("-", 16))
		}
		fmt.Println()

		fmt.Printf("%-10s", "总计")
		for _, state := range tcpStates {
			total := totalByState[state]
			if total > 0 {
				fmt.Printf(" | %-12d", total)
			} else {
				fmt.Printf(" | %-12s", "-")
			}
		}
		fmt.Println()
	}

	fmt.Printf("\n总连接数: %d\n", totalConnections)
}
