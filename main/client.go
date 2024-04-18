package main

import (
	"fmt"
	"net"
	"os"

	raft "github.com/yesmanchyk/go-raft/part2"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <address> <command1> <command2> ...\n", os.Args[0])
		return
	}
	addr := os.Args[1]
	n := 1
	node := 0
	commitChans := make([]chan raft.CommitEntry, 1)
	ready := make(chan interface{})
	peerIds := make([]int, 0)
	for p := 0; p < n; p++ {
		if p != node {
			peerIds = append(peerIds, p)
		}
	}
	commitChans[0] = make(chan raft.CommitEntry)
	raftServer := raft.NewServer(node, peerIds, ready, commitChans[0])
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil { 
		fmt.Printf("net.ResolveTCPAddr.Error=%v\n", err)
		return
	}
	fmt.Printf("TcpAddr=%v\n", tcpAddr);
	err = raftServer.ConnectToPeer(node, tcpAddr)
	if err != nil { 
		fmt.Printf("%v.ConnectToPeer.Error=%v\n", raftServer, err)
		return
	}
	args := raft.ReportArgs { Id: node }
	reply := raft.ReportReply { }
	fmt.Printf("ConsensusModule.RequestReport(%v)\n", args);
	err = raftServer.Call(node, "ConsensusModule.RequestReport", &args, &reply)
	if err != nil { 
		fmt.Printf("ConsensusModule.RequestReport.Error=%v\n", err)
		return
	}
	fmt.Printf("%d.ConsensusModule.Report.Reply=%v\n", node, reply);
	if reply.IsLeader {
		for _, cmd := range os.Args[2:] {
			args := raft.SubmitArgs { Command: cmd }
			var reply raft.SubmitReply
			fmt.Printf("ConsensusModule.SubmitCommand=%v\n", args)
			err = raftServer.Call(node, "ConsensusModule.SubmitCommand", &args, &reply)
			if err == nil {
				fmt.Printf("ConsensusModule.SubmitCommand.Reply=%v\n", reply)
			} else { 
				fmt.Printf("ConsensusModule.SubmitCommand.Error=%v\n", err)
			}	
		}
	}
}
