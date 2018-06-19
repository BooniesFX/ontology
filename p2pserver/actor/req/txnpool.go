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
