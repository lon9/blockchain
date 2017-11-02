package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/deckarep/golang-set"
)

type Blockchain struct {
	Chain               []Block
	currentTransactions []Transaction
	Nodes               mapset.Set
}

func NewBlockchain() *Blockchain {
	b := &Blockchain{
		Nodes: mapset.NewSet(),
	}
	b.NewBlock(100, "1")
	return b
}

func (b *Blockchain) NewBlock(proof int, previousHash string) (*Block, error) {

	if previousHash == "" {
		hash, err := b.hash(&b.Chain[len(b.Chain)-1])
		if err != nil {
			return nil, err
		}
		previousHash = hash
	}

	block := NewBlock(
		len(b.Chain),
		time.Now(),
		b.currentTransactions,
		proof,
		previousHash,
	)

	b.currentTransactions = nil
	b.Chain = append(b.Chain, *block)
	return block, nil
}

func (b *Blockchain) NewTransaction(sender, recipient string, amount int) int {
	b.currentTransactions = append(b.currentTransactions, *NewTransaction(sender, recipient, amount))
	return b.LastBlock().Index + 1
}

func (b *Blockchain) hash(block *Block) (string, error) {
	blockString, err := json.Marshal(block)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(blockString))), nil
}

func (b *Blockchain) LastBlock() *Block {
	return &b.Chain[len(b.Chain)-1]
}

func (b *Blockchain) ProofOfWork(lastProof int) int {
	proof := 0
	for {
		if b.validProof(lastProof, proof) {
			break
		}
		proof++
	}
	return proof
}

func (b *Blockchain) validProof(lastProof, proof int) bool {
	guess := string(lastProof) + string(proof)
	guessHash := fmt.Sprintf("%x", sha256.Sum256([]byte(guess)))
	return guessHash[:4] == "0000"
}

func (b *Blockchain) RegisterNode(addr string) error {
	u, err := url.Parse(addr)
	if err != nil {
		return err
	}
	b.Nodes.Add(u.Host)
	return nil
}

func (b *Blockchain) validChain(chain []Block) (bool, error) {
	lastBlock := chain[len(chain)-1]
	currentIndex := 1

	for {
		if currentIndex < len(chain) {
			break
		}
		block := chain[currentIndex]
		fmt.Print(lastBlock)
		fmt.Print(block)
		fmt.Print("\n--------\n")

		lashBlockHash, err := b.hash(&lastBlock)
		if err != nil {
			return false, err
		}
		if block.PreviousHash != lashBlockHash {
			return false, nil
		}

		if !b.validProof(lastBlock.Proof, block.Proof) {
			return false, nil
		}
		lastBlock = block
		currentIndex++
	}
	return true, nil
}

func (b *Blockchain) ResolveConflicts() (bool, error) {
	neighbors := b.Nodes
	var newChain []Block

	maxLength := len(b.Chain)

	it := neighbors.Iterator()
	for node := range it.C {
		res, err := http.Get(fmt.Sprintf("http://%s/chain", node))
		if err != nil {
			return false, err
		}
		if res.StatusCode == 200 {
			defer res.Body.Close()
			br, err := ioutil.ReadAll(res.Body)
			var r ChainResponse
			if err = json.Unmarshal(br, &r); err != nil {
				return false, err
			}
			length := r.Length
			chain := r.Chain
			ok, err := b.validChain(chain)
			if err != nil {
				return false, err
			}
			if length > maxLength && ok {
				maxLength = length
				newChain = chain
			}
		}
	}
	if len(newChain) != 0 {
		b.Chain = newChain
		return true, nil
	}
	return false, nil
}
