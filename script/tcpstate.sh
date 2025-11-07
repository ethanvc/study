#!/bin/bash

# TCP状态统计脚本 - 按服务端端口聚合分析
# 仅支持 macOS 系统
# 每2秒更新一次统计数据
# 兼容 bash 3.x（macOS 默认版本）
#
# 使用方法:
#   ./tcpstate.sh 80 443 8080        # 监控指定端口
#   ./tcpstate.sh                    # 交互式输入端口

# 检查是否为macOS系统
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo "错误: 此脚本仅支持 macOS 系统"
    exit 1
fi

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 要监控的端口列表（全局变量）
MONITOR_PORTS=""

# 清屏函数
clear_screen() {
    printf "\033c"
}

# 根据状态返回颜色代码
get_state_color() {
    local state=$1
    case $state in
        ESTABLISHED)
            echo -n "$GREEN"
            ;;
        LISTEN)
            echo -n "$BLUE"
            ;;
        TIME_WAIT)
            echo -n "$YELLOW"
            ;;
        CLOSE_WAIT)
            echo -n "$RED"
            ;;
        SYN_SENT|SYN_RCVD)
            echo -n "$CYAN"
            ;;
        FIN_WAIT_1|FIN_WAIT_2|CLOSING|LAST_ACK)
            echo -n "$YELLOW"
            ;;
        *)
            echo -n "$NC"
            ;;
    esac
}

# 统计TCP状态
analyze_tcp_stats() {
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}TCP状态统计报告${NC} (macOS) - ${timestamp}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${MAGENTA}监控端口: ${MONITOR_PORTS}${NC}"
    echo ""
    
    # 构建awk的端口过滤条件
    local port_filter=""
    for port in $MONITOR_PORTS; do
        if [[ -z "$port_filter" ]]; then
            port_filter="port == $port"
        else
            port_filter="$port_filter || port == $port"
        fi
    done
    
    # 使用awk处理netstat输出，只提取指定端口的连接
    # 输出格式: 端口 状态
    local stats=$(netstat -an -p tcp 2>/dev/null | awk -v filter="$port_filter" '
        /^tcp/ && $6 != "" && $6 != "Foreign" {
            # 提取本地地址（第4列）和状态
            local_addr = $4
            state = $6
            
            # 从地址中提取端口号（最后一个.后面的数字）
            if (match(local_addr, /\.([0-9]+)$/, arr)) {
                port = arr[1]
                # 只输出监控列表中的端口
                if ('"$port_filter"') {
                    print port, state
                }
            }
        }
    ')
    
    if [[ -z "$stats" ]]; then
        echo -e "${YELLOW}没有检测到指定端口的TCP连接${NC}"
        echo ""
        echo "按 Ctrl+C 退出监控"
        return
    fi
    
    # 使用sort和uniq统计每个端口+状态的组合
    local aggregated=$(echo "$stats" | sort -n | uniq -c | awk '{print $2, $3, $1}')
    # 格式: 端口 状态 数量
    
    echo -e "${YELLOW}端口\t\t状态分布${NC}"
    echo -e "${CYAN}───────────────────────────────────────────────────────────────${NC}"
    
    local total_connections=0
    local active_port_count=0
    
    # 遍历监控的每个端口（按输入顺序）
    for port in $MONITOR_PORTS; do
        # 获取该端口的所有状态和数量
        local port_stats=$(echo "$aggregated" | awk -v p="$port" '$1 == p {print $2, $3}')
        
        # 如果该端口没有连接，跳过
        if [[ -z "$port_stats" ]]; then
            continue
        fi
        
        ((active_port_count++))
        
        # 计算该端口的总连接数
        local port_total=$(echo "$port_stats" | awk '{sum += $2} END {print sum}')
        ((total_connections += port_total))
        
        echo -e "\n${MAGENTA}端口 $port${NC} (总连接数: ${GREEN}$port_total${NC})"
        
        # 输出该端口的各状态统计
        while read -r state count; do
            if [[ -n "$state" && -n "$count" ]]; then
                local color=$(get_state_color "$state")
                printf "  ${color}%-15s${NC}: %5d\n" "$state" "$count"
            fi
        done <<< "$port_stats"
    done
    
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}总连接数: $total_connections${NC} | ${MAGENTA}活跃端口数: $active_port_count${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
    echo ""
}

# 验证端口号是否合法
validate_port() {
    local port=$1
    if [[ "$port" =~ ^[0-9]+$ ]] && [ "$port" -ge 1 ] && [ "$port" -le 65535 ]; then
        return 0
    else
        return 1
    fi
}

# 获取要监控的端口列表
get_monitor_ports() {
    # 如果命令行提供了参数，使用命令行参数
    if [ $# -gt 0 ]; then
        for port in "$@"; do
            if validate_port "$port"; then
                MONITOR_PORTS="$MONITOR_PORTS $port"
            else
                echo -e "${RED}错误: 无效的端口号 '$port' (端口范围: 1-65535)${NC}"
                exit 1
            fi
        done
        # 去除首尾空格
        MONITOR_PORTS=$(echo "$MONITOR_PORTS" | xargs)
    else
        # 交互式输入
        echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
        echo -e "${GREEN}TCP状态监控工具${NC}"
        echo -e "${CYAN}═══════════════════════════════════════════════════════════════${NC}"
        echo ""
        echo -e "${YELLOW}请输入要监控的端口列表（空格分隔，例如: 80 443 8080）:${NC}"
        read -r input_ports
        
        if [[ -z "$input_ports" ]]; then
            echo -e "${RED}错误: 未输入端口${NC}"
            exit 1
        fi
        
        for port in $input_ports; do
            if validate_port "$port"; then
                MONITOR_PORTS="$MONITOR_PORTS $port"
            else
                echo -e "${RED}错误: 无效的端口号 '$port' (端口范围: 1-65535)${NC}"
                exit 1
            fi
        done
        # 去除首尾空格
        MONITOR_PORTS=$(echo "$MONITOR_PORTS" | xargs)
        
        echo ""
        echo -e "${GREEN}✓ 将监控以下端口: ${MONITOR_PORTS}${NC}"
        echo ""
        sleep 1
    fi
}

# 主程序
main() {
    # 获取端口列表
    get_monitor_ports "$@"
    
    # 确认有端口需要监控
    if [[ -z "$MONITOR_PORTS" ]]; then
        echo -e "${RED}错误: 没有有效的监控端口${NC}"
        exit 1
    fi
    
    echo "启动TCP状态监控..."
    echo "每2秒更新一次统计数据"
    echo ""
    sleep 1
    
    # 主循环
    while true; do
        analyze_tcp_stats
        sleep 2
    done
}

# 运行主程序
main "$@"

