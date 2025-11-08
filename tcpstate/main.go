package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/orderedmap/v3"
)

func main() {
	targetPosts := orderedmap.NewOrderedMap[int, bool]()
	for _, target := range os.Args[1:] {
		port, err := strconv.Atoi(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error parsing target port: %v\n", err)
			return
		}
		targetPosts.Set(port, true)
	}
	for {
		connections, err := AllConnections()
		if err != nil {
			panic(err)
		}
		statistic := getStatistic(connections, targetPosts)
		printResult(statistic, targetPosts)
		time.Sleep(2 * time.Second)
	}

}

var sStatusSLice = []string{
	"LISTEN", "ESTABLISHED", "SYN_SENT", "CLOSE_WAIT", "TIME_WAIT"}

func mustValidStatus(s string) {
	for _, status := range sStatusSLice {
		if s == status {
			return
		}
	}
	panic("invalid status: " + s)
}

func printResult(statistic map[int]map[string]int, targetPosts *orderedmap.OrderedMap[int, bool]) {
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%s\n", time.Now().Format(time.DateTime))
	fmt.Printf("%12s", "SVR_PORT")
	for _, val := range sStatusSLice {
		fmt.Printf("%12s", val)
	}
	fmt.Printf("\n")
	for port := range targetPosts.AllFromFront() {
		stat := statistic[port]
		fmt.Printf("%12d", port)
		for _, val := range sStatusSLice {
			fmt.Printf("%12s", getCountOr(stat, val, "-"))
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n\n")
}

func getCountOr(m map[string]int, key, placeholder string) string {
	if val, ok := m[key]; ok {
		return strconv.Itoa(val)
	} else {
		return placeholder
	}
}

func getStatistic(connections []*ConnectionInfo, targetPosts *orderedmap.OrderedMap[int, bool]) map[int]map[string]int {
	statistic := make(map[int]map[string]int)
	for _, connection := range connections {
		port := getPort(connection.LAddr)
		if !targetPosts.Has(port) {
			port = getPort(connection.RAddr)
		}
		if !targetPosts.Has(port) {
			continue
		}
		portStatistic := statistic[port]
		if portStatistic == nil {
			portStatistic = make(map[string]int)
			statistic[port] = portStatistic
		}
		portStatistic[connection.Status]++
	}
	return statistic
}

func getPort(addr net.Addr) int {
	switch realAddr := addr.(type) {
	case *net.TCPAddr:
		return realAddr.Port
	default:
		return 0
	}
}

type ConnectionInfo struct {
	LAddr  net.Addr
	RAddr  net.Addr
	Status string
}

func AllConnections() ([]*ConnectionInfo, error) {
	cmd := exec.Command("netstat", "-ptcp", "-nl")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	reg := regexp.MustCompile(`(\w+)\s+(\d+)\s+(\d+)\s+([\w.:]+)\s+([\w.:]+)\s+([\w_]+)`)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var result []*ConnectionInfo
	for scanner.Scan() {
		line := scanner.Text()
		parts := reg.FindStringSubmatch(line)
		if len(parts) != 7 {
			continue
		}
		laddrStr := parts[4]
		raddrStr := parts[5]
		status := parts[6]
		laddrStr = convertToStandardAddress(laddrStr)
		laddr, err := net.ResolveTCPAddr("tcp", laddrStr)
		if err != nil {
			return nil, err
		}
		raddrStr = convertToStandardAddress(raddrStr)
		raddr, err := net.ResolveTCPAddr("tcp", raddrStr)
		if err != nil {
			return nil, err
		}
		mustValidStatus(status)
		result = append(result, &ConnectionInfo{
			LAddr:  laddr,
			RAddr:  raddr,
			Status: status,
		})
	}
	return result, nil
}

func convertToStandardAddress(addr string) string {
	if strings.Contains(addr, ":") {
		idx := strings.LastIndexByte(addr, '.')
		if idx == -1 {
			return ""
		}
		return "[" + addr[:idx] + "]:" + addr[idx+1:]

	} else {
		idx := strings.LastIndexByte(addr, '.')
		if idx == -1 {
			return ""
		}
		return addr[:idx] + ":" + addr[idx+1:]
	}
}
