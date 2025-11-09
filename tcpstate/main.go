package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elliotchance/orderedmap/v3"
	"github.com/ethanvc/study/tcpstate/tcpstate"
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
	var oldConnections map[string]*tcpstate.ConnectionInfo
	for {
		connections, err := tcpstate.GetAllConnections()
		if err != nil {
			panic(err)
		}
		statistic, dynStatistic := getStatistic(oldConnections, connections, targetPosts)
		printResult(statistic, dynStatistic, targetPosts)
		oldConnections = connectionSliceToMap(connections)
		time.Sleep(2 * time.Second)
	}

}

func connectionSliceToMap(connections []*tcpstate.ConnectionInfo) map[string]*tcpstate.ConnectionInfo {
	m := make(map[string]*tcpstate.ConnectionInfo)
	for _, conn := range connections {
		m[conn.GetId()] = conn
	}
	return m
}

func printResult(statistic map[int]map[tcpstate.Status]int,
	dynStatistic map[int]map[tcpstate.Status]int,
	targetPosts *orderedmap.OrderedMap[int, bool]) {
	fmt.Println(strings.Repeat("=", 120))
	fmt.Printf("%s\n", time.Now().Format(time.DateTime))
	fmt.Printf("%8s", "SVR_PORT")
	for _, val := range tcpstate.AllStatus {
		fmt.Printf("%11s", val)
	}
	fmt.Printf("\n")
	for port := range targetPosts.AllFromFront() {
		stat := statistic[port]
		dynStat := dynStatistic[port]
		fmt.Printf("%8d", port)
		for _, val := range tcpstate.AllStatus {
			fmt.Printf("%11s", getCountOr(stat, val, "-"))
		}
		fmt.Printf("\n")
		fmt.Printf("%8s", " ")
		for _, val := range tcpstate.AllStatus {
			fmt.Printf("%11s", getCountOr(dynStat, val, "-"))
		}
		fmt.Printf("\n")
	}
	fmt.Printf("\n\n")
}

func getCountOr(m map[tcpstate.Status]int, key tcpstate.Status, placeholder string) string {
	if val, ok := m[key]; ok {
		return strconv.Itoa(val)
	} else {
		return placeholder
	}
}

func getStatistic(oldConnections map[string]*tcpstate.ConnectionInfo,
	connections []*tcpstate.ConnectionInfo, targetPosts *orderedmap.OrderedMap[int, bool]) (map[int]map[tcpstate.Status]int,
	map[int]map[tcpstate.Status]int) {
	statistic := make(map[int]map[tcpstate.Status]int)
	dynStatistic := make(map[int]map[tcpstate.Status]int)
	for port := range targetPosts.AllFromFront() {
		statistic[port] = make(map[tcpstate.Status]int)
		dynStatistic[port] = make(map[tcpstate.Status]int)
	}
	for _, connection := range connections {
		if !targetPosts.Has(connection.SvrPort) {
			continue
		}
		statistic[connection.SvrPort][connection.Status]++
		if oldConnections[connection.GetId()] == nil {
			dynStatistic[connection.SvrPort][connection.Status]++
		}
	}
	return statistic, dynStatistic
}
