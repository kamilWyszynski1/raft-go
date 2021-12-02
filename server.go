package raft

import "sync"

// LeaderCommunicationChan is a channel that will tell Server which Node is a leader.
type LeaderCommunicationChan chan int

// Server is a local server implementation that works with goroutines.
// Server also allows traffic from outside to Cluster.
type Server struct {
	Net     Network
	Cluster Cluster

	leader    *Node
	leaderMtx sync.Mutex
}

func NewServer(network Network, cluster Cluster) *Server {
	s := &Server{Net: network, Cluster: cluster}

	lcch := make(chan int)
	for _, n := range cluster.Nodes {
		n.SetLeaderCommunicationChan(lcch)
	}

	go func() {
		clusterMap := make(map[int]*Node)
		for _, node := range cluster.Nodes {
			clusterMap[node.id] = node
		}
		for {
			select {
			case leaderID := <-lcch:
				s.leaderMtx.Lock()
				s.leader = clusterMap[leaderID]
				s.leaderMtx.Unlock()
			}
		}
	}()
	return s
}

func (s *Server) RequestVote(from, to int, term int) (bool, int) {
	return s.Net.Net[to].RequestVote(from, term)
}

func (s *Server) AppendEntries(from, to, term int, entry Entry) (bool, int) {
	return s.Net.Net[to].AppendEntries(term, from, entry)
}

func (s *Server) ExternalRequest(v Value) {
	s.leaderMtx.Lock()
	defer s.leaderMtx.Unlock()
	s.leader.HandleExternalRequest(v)
}

func (s *Server) Print() {
	s.leaderMtx.Lock()
	defer s.leaderMtx.Unlock()
	s.leader.Print()
}
