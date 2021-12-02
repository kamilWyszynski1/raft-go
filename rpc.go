package raft

import "github.com/gofrs/uuid"

// EntryType is a Entry type.
type EntryType int8

const (
	UnknownEntry EntryType = iota
	Append                 // add entry to uncommitted entries.
	Commit                 // commits entry.
	Print                  // order node to print content.
)

type AddEntryType int8

const (
	UnknownAddEntry AddEntryType = iota
	Add
	Delete
)

// AddEntry implements Entry, first request to Node, add to uncommited map.
type AddEntry struct {
	ID           uuid.UUID
	Value        Value
	AddEntryType AddEntryType
}

func (a AddEntry) Type() EntryType {
	return Append
}

func (a AddEntry) Body() interface{} {
	return a
}

// CommitEntry implements Entry for committing AddEntry.
type CommitEntry struct {
	ID uuid.UUID
}

func (c CommitEntry) Type() EntryType {
	return Commit
}

func (c CommitEntry) Body() interface{} {
	return c.ID
}

type PrintEntry struct{}

func (p PrintEntry) Type() EntryType {
	return Print
}

func (p PrintEntry) Body() interface{} {
	return nil
}

// Entry is a raft entry interface.
type Entry interface {
	Type() EntryType
	Body() interface{}
}

// RPC is an interface that provides RPC functionalities across cluster.
type RPC interface {
	// RequestVote sends vote request.
	RequestVote(fromID, term int) (bool, int)
	// AppendEntries sends append entries request.
	AppendEntries(term, leaderID int, entry Entry) (success bool, replyTerm int)
}
