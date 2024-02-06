package main

import bft "github.com/SmartBFT-Go/consensus/pkg/types"

func (*Node) Sign(msg []byte) []byte {
	return nil
}

func (n *Node) SignProposal(bft.Proposal, []byte) *bft.Signature {
	return &bft.Signature{
		ID: n.id,
	}
}
