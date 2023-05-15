package db

import (
	"gorm.io/gorm"
	"sync/atomic"
)

type TxContext struct {
	TxUUID  string
	TxLen   int
	TxIndex int32
	Tx      *gorm.DB
}

func (c *TxContext) IncTxIndex() {
	atomic.AddInt32(&c.TxIndex, 1)
}
