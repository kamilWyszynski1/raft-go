package raft

type MockedNetwork map[int]RPC

type Network struct {
	Net MockedNetwork
}
