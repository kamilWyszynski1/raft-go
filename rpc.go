package raft

// RPC is an interface that provides RPC functionalities across cluster.
type RPC interface {
	// RequestVote sends vote request.
	RequestVote(fromID, term int) (bool, int)
	// AppendEntries sends append entries request.
	AppendEntries(term, leaderID int) (success bool, replyTerm int)
}
