package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/SmartBFT-Go/consensus/smartbftprotos"
	"github.com/golang/protobuf/proto"
	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type Communicator struct {
	nodeId            uint64
	mapNodes          map[uint64]*NodeInfo
	cachedHttpClients map[uint64]*http.Client
}

func (c Communicator) SendConsensus(targetID uint64, m *smartbftprotos.Message) {
	nodeInfo, ok := c.mapNodes[targetID]
	if !ok {
		log.Warnf(fmt.Sprintf("Node %d not found in configuration", targetID))
		return
	}
	// proto.message to bytes
	request, err := proto.Marshal(m)
	if err != nil {
		log.Warnf(fmt.Sprintf("Error marshaling proto.message: %s", err))
		return
	}
	http3Client := c.getOrCreateClient(targetID)
	// send request to node
	log.Debugf("node %d sent consensus to node %d address=%s", c.nodeId, targetID, nodeInfo.Address)
	go func() {
		r := bytes.NewBuffer(request)
		req, err := http3Client.Post("https://"+nodeInfo.Address+"/consensus?id="+strconv.FormatUint(c.nodeId, 10), "application/octet-stream", r)
		if err != nil {
			log.Warnf(fmt.Sprintf("Error sending request to node: %s", err))
			return
		}
		req.Body.Close()
		log.Debugf("Node %d sent consensus to node %d", c.nodeId, targetID)
	}()
}

func (c Communicator) SendTransaction(targetID uint64, request []byte) {
	nodeInfo, ok := c.mapNodes[targetID]
	if !ok {
		log.Warnf(fmt.Sprintf("Node %d not found in configuration", targetID))
		return
	}
	msg := &FwdMessage{
		Payload: request,
		Sender:  c.nodeId,
	}
	// proto.message to bytes
	bodyReq, err := proto.Marshal(msg)
	if err != nil {
		log.Warnf(fmt.Sprintf("Error marshaling proto.message: %s", err))
		return
	}
	http3Client := c.getOrCreateClient(targetID)
	// send request to node
	log.Infof("node %d sending transaction to node %d address=%s", c.nodeId, targetID, nodeInfo.Address)
	go func() {
		req, err := http3Client.Post("https://"+nodeInfo.Address+"/transaction?id="+strconv.FormatUint(c.nodeId, 10), "application/octet-stream", bytes.NewBuffer(bodyReq))
		if err != nil {
			log.Warnf(fmt.Sprintf("Error sending request to node: %s", err))
			return
		}
		req.Body.Close()
		log.Infof("Node %d sent transaction to node %d", c.nodeId, targetID)
	}()
}

func (c Communicator) getOrCreateClient(targetID uint64) *http.Client {
	http3Client, ok := c.cachedHttpClients[targetID]
	if ok {
		return http3Client
	}
	rt := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"quic-echo-example"},
		},
		DisableCompression: true,
		QuicConfig:         &quic.Config{MaxIdleTimeout: 10 * time.Second},
	}
	http3Client = &http.Client{
		Transport: rt,
	}
	c.cachedHttpClients[targetID] = http3Client
	return http3Client
}

func (c Communicator) Nodes() []uint64 {
	nodes := make([]uint64, 0, len(c.mapNodes))
	for id := range c.mapNodes {
		nodes = append(nodes, id)
	}

	return nodes
}
