package tcpstate

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func GetAllConnections() ([]*ConnectionInfo, error) {
	cmd := exec.Command("netstat", "-ptcp", "-anl")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	reg := regexp.MustCompile(`(\w+)\s+(\d+)\s+(\d+)\s+([\w.:*]+)\s+([\w.:*]+)\s+([\w_]+)`)
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
		status := ToStatus(parts[6])
		status.MustValid()
		var conn ConnectionInfo
		conn.Status = status
		conn.ClientAddr, conn.ClientPort, err = parseAddr(laddrStr)
		if err != nil {
			return nil, err
		}
		conn.SvrAddr, conn.SvrPort, err = parseAddr(raddrStr)
		if err != nil {
			return nil, err
		}
		if conn.ClientPort == 0 && conn.SvrPort == 0 {
			continue
		}

		result = append(result, &conn)
	}
	fixAddrs(result)
	return result, nil
}

func fixAddrs(connections []*ConnectionInfo) {
	svrPorts := make(map[int]bool)
	for _, conn := range connections {
		if conn.Status == StatusListen {
			svrPorts[conn.ClientPort] = true
		}
	}
	for _, conn := range connections {
		if svrPorts[conn.ClientPort] {
			conn.SvrAddr, conn.ClientAddr = conn.ClientAddr, conn.SvrAddr
			conn.SvrPort, conn.ClientPort = conn.ClientPort, conn.SvrPort
		}
	}
}

func parseAddr(addrStr string) (string, int, error) {
	var err error
	idx := strings.LastIndexByte(addrStr, '.')
	if idx == -1 {
		return "", 0, errors.New("invalid address: " + addrStr)
	}
	ipStr := addrStr[:idx]
	portStr := addrStr[idx+1:]
	var port int
	if portStr != "*" {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return "", 0, fmt.Errorf("invalid port: %s", portStr)
		}
	}

	return ipStr, port, nil
}
