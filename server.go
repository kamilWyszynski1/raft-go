package raft

// LeaderCommunicationChan is a channel that will tell Server which Node is a leader.
type LeaderCommunicationChan chan int

// Server is a local server implementation that works with goroutines.
// Server also allows traffic from outside to Cluster.
type Server struct {
	Net     Network
	Cluster Cluster

	leader *Node
}

func NewServer(network Network, cluster Cluster) *Server {
	s := &Server{Net: network, Cluster: cluster}

	lcch := make(chan int)
	for _, n := range cluster.Nodes {
		n.SetLeaderCommunicationChan(lcch)
	}
	return s
}

func (s Server) RequestVote(from, to int, term int) (bool, int) {
	return s.Net.Net[to].RequestVote(from, term)
}

func (s Server) AppendEntries(from, to, term int) (bool, int) {
	return s.Net.Net[to].AppendEntries(term, from)
}
