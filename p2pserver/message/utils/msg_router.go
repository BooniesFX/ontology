/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

package utils

import (
	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/common/config"
	"github.com/ontio/ontology/common/log"
	msgCommon "github.com/ontio/ontology/p2pserver/common"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
)

type MsgPack struct {
	data   *types.MsgPayload
	handle MessageHandler
	p2p    p2p.P2P
	pid    *actor.PID
}

// MessageHandler defines the unified api for each net message
type MessageHandler func(data *types.MsgPayload, p2p p2p.P2P, pid *actor.PID, args ...interface{})

// MessageRouter mostly route different message type-based to the
// related message handler
type MessageRouter struct {
	syncDispatcher *Dispatcher // Dispatch all msg from sync port
	consDispathcer *Dispatcher // Dispatch all msg from cons port
}

// NewMsgRouter returns a message router object
func NewMsgRouter(p2p p2p.P2P) *MessageRouter {
	msgRouter := &MessageRouter{}
	msgRouter.init(p2p)
	return msgRouter
}

// init initializes the message router's attributes
func (this *MessageRouter) init(p2p p2p.P2P) {
	poolsize := config.DefConfig.P2PNode.MaxRoutinePoolSize
	this.syncDispatcher = &Dispatcher{
		Source:   p2p.GetMsgChan(false),
		MsgCount: poolsize,
		MsgPool:  make(chan chan *MsgPack, poolsize),
		QuitPool: make(chan chan bool, poolsize),
		Handlers: make(map[string]MessageHandler),
		P2P:      p2p,
	}

	this.consDispathcer = &Dispatcher{
		Source:   p2p.GetMsgChan(true),
		MsgCount: poolsize,
		MsgPool:  make(chan chan *MsgPack, poolsize),
		QuitPool: make(chan chan bool, poolsize),
		Handlers: make(map[string]MessageHandler),
		P2P:      p2p,
	}

	// Register sync message handler
	this.syncDispatcher.RegisterMsgHandler(msgCommon.VERSION_TYPE, VersionHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.VERACK_TYPE, VerAckHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.GetADDR_TYPE, AddrReqHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.ADDR_TYPE, AddrHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.PING_TYPE, PingHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.PONG_TYPE, PongHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.GET_HEADERS_TYPE, HeadersReqHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.HEADERS_TYPE, BlkHeaderHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.INV_TYPE, InvHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.GET_DATA_TYPE, DataReqHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.BLOCK_TYPE, BlockHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.CONSENSUS_TYPE, ConsensusHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.NOT_FOUND_TYPE, NotFoundHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.TX_TYPE, TransactionHandle)
	this.syncDispatcher.RegisterMsgHandler(msgCommon.DISCONNECT_TYPE, DisconnectHandle)

	// Register cons message handler
	this.consDispathcer.RegisterMsgHandler(msgCommon.VERSION_TYPE, VersionHandle)
	this.consDispathcer.RegisterMsgHandler(msgCommon.VERACK_TYPE, VerAckHandle)
	this.consDispathcer.RegisterMsgHandler(msgCommon.CONSENSUS_TYPE, ConsensusHandle)
	this.consDispathcer.RegisterMsgHandler(msgCommon.DISCONNECT_TYPE, DisconnectHandle)

}

// SetPID sets p2p actor
func (this *MessageRouter) SetPID(pid *actor.PID) {
	this.syncDispatcher.Pid = pid
	this.consDispathcer.Pid = pid
}

// Start starts the loop to handle the message from the network
func (this *MessageRouter) Start() {
	this.syncDispatcher.Run()
	this.consDispathcer.Run()
	log.Info("MessageRouter start to parse p2p message...")
}

// Stop stops the message router's loop
func (this *MessageRouter) Stop() {
	this.syncDispatcher.Halt()
	this.consDispathcer.Halt()
}
