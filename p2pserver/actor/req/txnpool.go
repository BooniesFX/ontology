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

package req

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ontio/ontology-eventbus/actor"
	"github.com/ontio/ontology/core/types"
	p2pcommon "github.com/ontio/ontology/p2pserver/common"
)

var TxCnt uint64
var TxCntLatest uint64

const txnPoolReqTimeout = p2pcommon.ACTOR_TIMEOUT * time.Second

var txnPoolPid *actor.PID

func SetTxnPoolPid(txnPid *actor.PID) {
	txnPoolPid = txnPid
}

var DefTxnPid *actor.PID

type TxnPoolActor struct {
	props *actor.Props
}

func NewTxnPoolActor() *TxnPoolActor {
	return &TxnPoolActor{}
}

func (self *TxnPoolActor) Start() *actor.PID {
	self.props = actor.FromProducer(func() actor.Actor { return self })
	var err error
	DefTxnPid, err = actor.SpawnNamed(self.props, "TxnPoolActor")
	if err != nil {
		panic(fmt.Errorf("TxnPoolActor SpawnNamed error:%s", err))
	}
	return DefTxnPid
}

func (self *TxnPoolActor) Receive(ctx actor.Context) {
	switch msg := ctx.Message().(type) {
	case *actor.Started:
	case *actor.Stop:
	case *tc.TxReq:
		AddTransaction(msg.Tx)
	case *tc.GetTxnReq:
		sender := ctx.Sender()
		if sender != nil {
			sender.Request(&tc.GetTxnRsp{Txn: nil},
				ctx.Self())
		}
	default:
		log.Warnf("TxnPoolActor cannot deal with type: %v %v", msg, reflect.TypeOf(msg))
	}
}

//add txn to txnpool
func AddTransaction(transaction *types.Transaction) {
	atomic.AddUint64(&(TxCnt), 1)
}

func PrintTxnInfo() {
	txnPerSnd := TxCnt - TxCntLatest
	TxCntLatest = TxCnt
	fmt.Printf("total txn count %d,TPS = %d/s\n", TxCnt, txnPerSnd)
}

func LoopPrintActorInfo() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			PrintTxnInfo()
		}
	}
}
