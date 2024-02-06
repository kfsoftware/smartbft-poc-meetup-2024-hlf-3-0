package main

import (
	"fmt"
	smart "github.com/SmartBFT-Go/consensus/pkg/api"
	"github.com/SmartBFT-Go/consensus/pkg/metrics/disabled"
	"github.com/SmartBFT-Go/consensus/pkg/wal"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
	"time"
)

func main() {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	numNodes := 4
	networkOpts := NetworkOptions{
		NumNodes:     numNodes,
		BatchSize:    10,
		BatchTimeout: 2 * time.Second,
	}
	network := make(map[int]map[int]chan proto.Message)

	chains := make(map[int]*Chain)

	for id := 1; id <= networkOpts.NumNodes; id++ {
		network[id] = make(map[int]chan proto.Message)
		for i := 1; i <= networkOpts.NumNodes; i++ {
			network[id][i] = make(chan proto.Message, 128)
		}
	}

	mapNodes := make(map[uint64]*NodeInfo)
	for id := 1; id <= networkOpts.NumNodes; id++ {
		mapNodes[uint64(id)] = &NodeInfo{
			ID:         uint64(id),
			Address:    fmt.Sprintf("localhost:%d", 10000+id),
			OpsAddress: fmt.Sprintf("localhost:%d", 20000+id),
		}
	}

	for id := 1; id <= networkOpts.NumNodes; id++ {
		log.Infof("Initializing node %d", id)
		address := mapNodes[uint64(id)].Address
		opsAddress := mapNodes[uint64(id)].OpsAddress
		met := &disabled.Provider{}
		walMet := wal.NewMetrics(met, "label1")
		bftMet := smart.NewMetrics(met, "label1")
		walDir := fmt.Sprintf("./data/node%d", id)
		blocksDir := fmt.Sprintf("./blocks/node%d", id)
		logrusLogger := log.New()

		logrusLogger.SetLevel(log.InfoLevel)
		chain := NewChain(
			uint64(id),
			address,
			opsAddress,
			mapNodes,
			logrusLogger.WithField("nodeId", id).WithField("address", address),
			walMet,
			bftMet,
			networkOpts,
			walDir,
			blocksDir,
		)
		go func() {
			nodeID := id
			_ = nodeID
			for {
				block := chain.Listen()
				_ = block
				//log.Infof("Node: %d block: %v", nodeID, block)
			}
		}()
		chains[id] = chain
	}
	select {}
}
