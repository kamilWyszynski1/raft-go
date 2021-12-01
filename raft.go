package raft

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	heartBeatTick   = 2 * time.Second // heart beat will be sent with that frequency.
	runElectionTick = 1 * time.Second // run election tick frequency.
	timeoutDuration = 5 * time.Second // timeout duration.
)

// NodeState represents state of Node in Raft cluster.
type NodeState int8

const (
	// Follower is passive and merely responds to the requests of the leader or candidates
	Follower NodeState = iota
	// Candidate - when there’s no leader in the Raft cluster, any Follower can become a Candidate
	Candidate
	// Leader is fully responsible for managing the Raft cluster and it handles all the client commands coming in.
	Leader
)

// String returns readable NodeState.
func (n NodeState) String() string {
	switch n {
	case Follower:
		return "Follower"
	case Candidate:
		return "Candidate"
	case Leader:
		return "Leader"
	default:
		return "Unknown"

	}
}

// Cluster is a raft cluster.
type Cluster struct {
	Nodes []*Node
}

// Node represents raft node.
type Node struct {
	logger     logrus.FieldLogger
	mtx        sync.Mutex
	id         int
	peerIds    []int
	server     *Server
	logContent LinkedList

	state              NodeState
	lastHearBeat       time.Time
	electionResetEvent time.Time

	timeout  time.Duration
	term     int
	votedFor int

	lcch LeaderCommunicationChan
}

func NewNode(id int, logger *logrus.Logger, peerIds []int) *Node {
	return &Node{
		logger:     logger.WithField("id", id),
		id:         id,
		peerIds:    peerIds,
		state:      Follower,
		timeout:    timeoutDuration,
		logContent: NewSimpleLL(),
	}
}

func (n *Node) SetServer(server *Server) {
	n.server = server
}

func (n *Node) SetLeaderCommunicationChan(lcch LeaderCommunicationChan) {
	n.lcch = lcch
}

func (n *Node) AppendEntries(term, leaderID int) (success bool, replyTerm int) {
	logger := n.logger.WithFields(logrus.Fields{
		"method":    "AppendEntries",
		"term":      term,
		"leader_id": leaderID,
	})

	n.mtx.Lock()
	defer n.mtx.Unlock()
	currentTerm := n.term
	logger.WithField("current_term", currentTerm)

	logger.Info("appending entries")

	if term > currentTerm {
		n.becomeFollower(term)
	}

	if term == currentTerm {
		if n.state != Follower {
			n.becomeFollower(term)
		}
		n.electionResetEvent = time.Now()
		success = true
	}
	replyTerm = n.term
	return
}

func (n *Node) RequestVote(fromID, term int) (bool, int) {
	logger := n.logger.WithFields(logrus.Fields{
		"method":  "RequestVote",
		"from_id": fromID,
	})
	n.mtx.Lock()
	defer n.mtx.Unlock()
	logger.Info("voting")

	if term > n.term {
		n.becomeFollower(term)
	}

	if term == n.term && (n.votedFor == -1 || n.votedFor == fromID) {
		logger.Infof("voting for %d", fromID)

		n.electionResetEvent = time.Now()
		n.votedFor = fromID
		return true, n.term
	}

	logger.Info("not voting")
	return false, n.term
}

func (n *Node) becomeFollower(term int) {
	n.logger.WithField("method", "becomeFollower").WithField("term", term).Info("becoming follower")
	n.state = Follower
	n.term = term
	n.votedFor = -1
	n.electionResetEvent = time.Now()

	go n.runElectionTime()
}

func (n *Node) Start() {
	n.mtx.Lock()
	n.electionResetEvent = time.Now()
	n.mtx.Unlock()
	go func() {
		n.runElectionTime()
	}()
}

func (n *Node) becomeLeader() {
	n.logger.WithField("method", "becomeLeader").Warn("becoming leader")
	n.state = Leader
	n.lcch <- n.id

	go func() {
		ticker := time.NewTicker(heartBeatTick)
		for {
			select {
			case <-ticker.C:
				n.mtx.Lock()
				state := n.state
				n.mtx.Unlock()

				if state != Leader {
					return
				}
				n.sendHeartBeat()
			}
		}
	}()
}

func (n *Node) runElectionTime() {
	logger := n.logger.WithField("method", "runElectionTime")
	logrus.Info("run election time")
	ticker := time.NewTicker(runElectionTick)

	n.mtx.Lock()
	termStarted := n.term
	n.mtx.Unlock()
	for {
		select {
		case <-ticker.C:
			n.mtx.Lock()
			electionResetEvent := n.electionResetEvent
			state := n.state
			currentTerm := n.term
			n.mtx.Unlock()

			if state != Candidate && state != Follower {
				logger.Info("bailing out")
				return
			}

			if termStarted != currentTerm {
				logger.Info("current term is different than term started")
				return
			}

			if time.Since(electionResetEvent) > n.timeout {
				n.startElection()
				return
			}
		}
	}
}

func (n *Node) startElection() {
	logger := n.logger.WithField("method", "startElection")
	logger.Info("starting election")
	n.mtx.Lock()
	n.state = Candidate
	n.term += 1
	currentTerm := n.term
	n.mtx.Unlock()

	votesReceived := 1

	for _, to := range n.peerIds {
		go func(to int) {
			vote, term := n.server.RequestVote(n.id, to, currentTerm)
			n.mtx.Lock()
			defer n.mtx.Unlock()

			n.term += 1
			if n.state != Candidate {
				logger.WithField("current_state", n.state).Info("node is no longer candidate")
				return
			}

			if term > currentTerm {
				logger.Info("term out of date")
				n.becomeFollower(term)
				return
			}

			if term == currentTerm {
				if vote {
					votesReceived += 1
					if 2*votesReceived > len(n.peerIds)+1 {
						n.becomeLeader()
						return
					}
				}
			}
			return
		}(to)

	}
	go n.runElectionTime() // run new election if this is not succeed.
}

func (n *Node) sendHeartBeat() {
	n.logger.WithField("method", "sendHeartBeat").Info("beat")

	n.mtx.Lock()
	savedCurrentTerm := n.term
	n.mtx.Unlock()

	for _, to := range n.peerIds {
		go func(to int) {
			_, replyTerm := n.server.AppendEntries(n.id, to, savedCurrentTerm)
			n.mtx.Lock()
			defer n.mtx.Unlock()
			if replyTerm > savedCurrentTerm {
				n.becomeFollower(replyTerm)
			}
		}(to)
	}
}
