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
	"github.com/ontio/ontology/common/log"
	"github.com/ontio/ontology/p2pserver/message/types"
	"github.com/ontio/ontology/p2pserver/net/protocol"
)

//concurrent Msger
type Msger struct {
	MsgPool  chan chan *MsgPack //pubic msg pool
	MsgBag   chan *MsgPack      //msger`s own msg chan
	QuitPool chan chan bool     //public quit signal
	quit     chan bool          //quit signal
}

func NewMsger(msgpool chan chan *MsgPack, quitPool chan chan bool) *Msger {
	return &Msger{
		MsgPool:  msgpool,
		MsgBag:   make(chan *MsgPack),
		QuitPool: quitPool,
		quit:     make(chan bool),
	}
}

func (m *Msger) Start() {
	go func() {
		for {
			m.MsgPool <- m.MsgBag
			select {
			case pack := <-m.MsgBag:
				pack.handle(pack.data, pack.p2p, pack.pid)
			case <-m.quit:
				return
			}
		}
	}()
}

type Dispatcher struct {
	Source   chan *types.MsgPayload
	MsgCount uint
	MsgPool  chan chan *MsgPack //pubic msg pool
	QuitPool chan chan bool
	Handlers map[string]MessageHandler
	P2P      p2p.P2P    // Refer to the p2p network
	Pid      *actor.PID // P2P actor
}

// RegisterMsgHandler registers msg handler with the msg type
func (this *Dispatcher) RegisterMsgHandler(key string,
	handler MessageHandler) {
	this.Handlers[key] = handler
}

// UnRegisterMsgHandler un-registers the msg handler with
// the msg type
func (this *Dispatcher) UnRegisterMsgHandler(key string) {
	delete(this.Handlers, key)
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case msg := <-d.Source:
			go func() {
				msgType := msg.Payload.CmdType()
				handler, ok := d.Handlers[msgType]
				if ok {
					msgbag := <-d.MsgPool
					pack := &MsgPack{
						data:   msg,
						handle: handler,
						p2p:    d.P2P,
						pid:    d.Pid,
					}
					msgbag <- pack
				} else {
					log.Info("unknown message handler for the msg: ",
						msgType)
				}
			}()
		}
	}
}

func (d *Dispatcher) Run() {
	for i := uint(0); i < d.MsgCount; i++ {
		msger := NewMsger(d.MsgPool, d.QuitPool)
		msger.Start()
	}
	go d.dispatch()
}

func (d *Dispatcher) Halt() {
	for i := uint(0); i < d.MsgCount; i++ {
		quit, ok := <-d.QuitPool
		if ok {
			quit <- true
		}
	}
}
