package tcpstate

import "fmt"

type ConnectionInfo struct {
	ClientAddr string
	ClientPort int
	SvrAddr    string
	SvrPort    int
	Status     Status
}

func (ci *ConnectionInfo) GetId() string {
	return fmt.Sprintf("%s:%d/%s:%d", ci.ClientAddr, ci.ClientPort, ci.SvrAddr, ci.SvrPort)
}

type Status int

const (
	statusStart Status = iota
	StatusListen
	StatusFinWait1
	StatusFinWait2
	StatusEstablish
	StatusSyncSent
	StatusCloseWait
	StatusTimeWait
	StatusClosed
	StatusClosing
	StatusSyncRcvd
	StatusLastAck
	statusEnd
)

func (s Status) String() string {
	switch s {
	case statusStart, statusEnd:
		return "INVALID"
	case StatusListen:
		return "LISTEN"
	case StatusFinWait1:
		return "FIN_WAIT1"
	case StatusFinWait2:
		return "FIN_WAIT2"
	case StatusEstablish:
		return "ESTABLISH"
	case StatusSyncSent:
		return "SYNC_SENT"
	case StatusCloseWait:
		return "CLOSE_WAIT"
	case StatusTimeWait:
		return "TIME_WAIT"
	case StatusClosed:
		return "CLOSED"
	case StatusClosing:
		return "CLOSING"
	case StatusSyncRcvd:
		return "SYNC_RCVD"
	case StatusLastAck:
		return "LASK_ACK"
	}
	return "INVALID"
}

func ToStatus(s string) Status {
	switch s {
	case "LISTEN":
		return StatusListen
	case "FIN_WAIT_1":
		return StatusFinWait1
	case "ESTABLISHED":
		return StatusEstablish
	case "CLOSE_WAIT":
		return StatusCloseWait
	case "TIME_WAIT":
		return StatusTimeWait
	case "SYN_SENT":
		return StatusSyncSent
	case "CLOSED":
		return StatusClosed
	case "FIN_WAIT_2":
		return StatusFinWait2
	case "LAST_ACK":
		return StatusLastAck
	case "SYN_RCVD":
		return StatusSyncRcvd
	case "CLOSING":
		return StatusClosing
	default:
		return statusStart
	}
}

func (s Status) MustValid() {
	if s <= statusStart || s >= statusEnd {
		panic(fmt.Sprintf("invalid status %d", s))
	}
}

var AllStatus []Status

func init() {
	for i := statusStart + 1; i < statusEnd; i++ {
		AllStatus = append(AllStatus, i)
	}
}
