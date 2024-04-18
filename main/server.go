package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	raft "github.com/yesmanchyk/go-raft/part2"
)

type SingleServer struct {
	mu sync.Mutex

	// cluster is a list of all the raft servers participating in a cluster.
	server *raft.Server

	// commitChans has a channel per server in cluster with the commit channel for
	// that server.
	commitChans []chan raft.CommitEntry

	// commits at index i holds the sequence of commits made by server i so far.
	// It is populated by goroutines that listen on the corresponding commitChans
	// channel.
	commits [][]raft.CommitEntry

	n int
}

func NewSingleServer(cluster []string, node int) *SingleServer {
	n := len(cluster)
	commitChans := make([]chan raft.CommitEntry, 1)
	commits := make([][]raft.CommitEntry, 1)
	ready := make(chan interface{})

	// Create all Servers in this cluster, assign ids and peer ids.
	peerIds := make([]int, 0)
	for p := 0; p < n; p++ {
		if p != node {
			peerIds = append(peerIds, p)
		}
	}

	commitChans[0] = make(chan raft.CommitEntry)
	raftServer := raft.NewServer(node, peerIds, ready, commitChans[0])
	raftServer.ServeAt(cluster[node])

	// Connect all peers to each other.
	for j := 0; j < n; j++ {
		if node != j {
			tcpAddr, err := net.ResolveTCPAddr("tcp", cluster[j])
			if err != nil {
				continue
			}
			for {
				err = raftServer.ConnectToPeer(j, tcpAddr)
				if err == nil {
					break
				}
				log.Printf("Failed to connect to %s due to %v, retrying in 5 seconds", cluster[j], err)
				sleepMs(5000)
			}
		}
	}
	close(ready)

	h := &SingleServer{
		server:     raftServer,
		commitChans: commitChans,
		commits:     commits,
		n:           n,
	}
	return h
}

// Shutdown shuts down all the servers in the harness and waits for them to
// stop running.
func (h *SingleServer) Shutdown() {
	h.server.DisconnectAll()
	h.server.Shutdown()
	for i := 0; i < h.n; i++ {
		close(h.commitChans[i])
	}
}

func sleepMs(n int) {
	time.Sleep(time.Duration(n) * time.Millisecond)
}

func main() {
	cluster := os.Getenv("RAFT_CLUSTER")
	if cluster == "" {
		fmt.Printf("RAFT_CLUSTER=%v, please run the following line:\n", cluster)
		fmt.Println("export RAFT_CLUSTER=localhost:12341,localhost:12342,localhost:12343")
		return
	}
	addrs := strings.Split(cluster, ",")
	fmt.Printf("RAFT_CLUSTER=%v\n", addrs)
	n, err := strconv.Atoi(os.Getenv("RAFT_ID"))
	if err != nil {
		fmt.Printf("RAFT_ID parsing error %v, please run the following line:\n", err)
		fmt.Println("export RAFT_ID=0")
		return
	}
	fmt.Printf("RAFT_ID=%v\n", n)
	if len(addrs) <= n {
		fmt.Printf("RAFT_ID should be less than %v\n", len(addrs))
		return;
	}

	h := NewSingleServer(addrs, n)
	defer func() {
		fmt.Printf("Shutting down %v\n", h)
		h.Shutdown()
	}()

	fmt.Printf("Started %v\n", h)
	fmt.Printf("Press any key to exit...\n", h)
	var b []byte = make([]byte, 1)
	os.Stdin.Read(b)
	fmt.Println("I got the byte", b, "("+string(b)+")")
}
