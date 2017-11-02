package blockchain

type ChainResponse struct {
	Chain  []Block `json:"chain"`
	Length int     `json:"length"`
}
