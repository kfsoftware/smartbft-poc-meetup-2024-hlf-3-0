package main

import (
	smart "github.com/SmartBFT-Go/consensus/pkg/api"

	"github.com/SmartBFT-Go/consensus/pkg/wal"
)

type Chain struct {
	deliverChan <-chan *Block
	node        *Node
}

func NewChain(
	id uint64,
	address string,
	opsAddress string,
	mapNodes map[uint64]*NodeInfo,
	logger smart.Logger,
	walmet *wal.Metrics,
	bftmet *smart.Metrics,
	opts NetworkOptions,
	walDir string,
	blocksDir string,
) *Chain {
	deliverChan := make(chan *Block)
	node := NewNode(
		id,
		address,
		opsAddress,
		mapNodes,
		deliverChan,
		logger,
		walmet,
		bftmet,
		opts,
		walDir,
		blocksDir,
	)
	return &Chain{
		node:        node,
		deliverChan: deliverChan,
	}
}

func (chain *Chain) Listen() Block {
	block := <-chain.deliverChan
	return *block
}

func (chain *Chain) Order(txn Transaction) error {
	return chain.node.consensus.SubmitRequest(txn.ToBytes())
}
