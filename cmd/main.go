package main

import (
	"fmt"
	raft "raft-go"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()

	ids := []int{0, 1, 2, 3, 4}
	var nodes []*raft.Node
	net := raft.Network{Net: map[int]raft.RPC{}}
	for index, id := range ids {
		peers := removeIndex(ids, index)
		fmt.Println(peers, ids)
		nodes = append(nodes, raft.NewNode(id, logger, peers))
		net.Net[id] = nodes[index]
	}

	server := raft.NewServer(net)

	for _, node := range nodes {
		node.SetServer(server)
	}

	for _, node := range nodes {
		go node.Start()
	}

	time.Sleep(time.Minute)
}

func removeIndex(s []int, index int) []int {
	next := make([]int, 0, len(s))
	for i, v := range s {
		if i != index {
			next = append(next, v)
		}
	}
	return next
}
