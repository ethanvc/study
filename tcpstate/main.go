package main

import (
	"bufio"
	"bytes"
	"net"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	AllConnections()

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
		result = append(result, &ConnectionInfo{
			LAddr: laddr,
			RAddr: raddr,
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
