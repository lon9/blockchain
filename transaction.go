package blockchain

type Transaction struct {
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Amount    int    `json:"amount"`
}

func NewTransaction(sender, recipient string, amount int) *Transaction {
	return &Transaction{
		sender,
		recipient,
		amount,
	}
}
