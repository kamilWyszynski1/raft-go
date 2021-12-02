package main

import (
	"bufio"
	"fmt"
	"os"
	raft "raft-go"
	"strings"

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

	server := raft.NewServer(net, raft.Cluster{Nodes: nodes})

	for _, node := range nodes {
		node.SetServer(server)
	}

	for _, node := range nodes {
		go node.Start()
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if text == "exit" {
			return
		}

		if text == "print" {
			server.Print()
		}

		split := strings.Split(text, " ")
		if split[0] == "append" {
			v := raft.Value(split[2])
			switch split[1] {
			case "add":
				server.ExternalRequest(v)
			case "delete":
				fmt.Print("implement")
			}
		}
	}
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
