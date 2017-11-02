package blockchain

import "time"

type Block struct {
	Index        int           `json:"index"`
	TimeStamp    time.Time     `json:"timeStamp"`
	Transactions []Transaction `json:"transactions"`
	Proof        int           `json:"proof"`
	PreviousHash string        `json:"previousHash"`
}

func NewBlock(index int, timeStamp time.Time, transactions []Transaction, proof int, previousHash string) *Block {
	return &Block{
		index,
		timeStamp,
		transactions,
		proof,
		previousHash,
	}
}
