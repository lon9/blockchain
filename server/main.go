package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/Rompei/blockchain"
	"github.com/labstack/echo"
	"github.com/satori/go.uuid"
)

type TransactionResponse struct {
	Message string `json:"message"`
}

type MineResponse struct {
	Message      string                   `json:"message"`
	Index        int                      `json:"index"`
	Transactions []blockchain.Transaction `json:"transactions"`
	Proof        int                      `json:"proof"`
	PreviousHash string                   `json:"previousHash"`
}

type ChainResponse struct {
	Chain  []blockchain.Block `json:"chain"`
	Length int                `json:"length"`
}

type RegisterBody struct {
	Nodes []string `json:"nodes"`
}

type RegisterResponse struct {
	Message    string        `json:"message"`
	TotalNodes []interface{} `json:"totalNodes"`
}

type ResolveResponse struct {
	Message string             `json:"message"`
	Chain   []blockchain.Block `json:"chain"`
}

func main() {

	var port string

	flag.StringVar(&port, "p", "5000", "Port number")
	flag.Parse()

	nodeIdentifier := strings.Replace(uuid.NewV4().String(), "-", "", -1)

	bc := blockchain.NewBlockchain()

	e := echo.New()
	e.POST("/transactions/new", func(c echo.Context) error {
		transaction := new(blockchain.Transaction)
		if err := c.Bind(transaction); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		index := bc.NewTransaction(transaction.Sender, transaction.Recipient, transaction.Amount)
		res := &TransactionResponse{
			Message: fmt.Sprintf("Transaction was added in block %d", index),
		}
		return c.JSON(http.StatusCreated, res)
	})
	e.GET("/mine", func(c echo.Context) error {
		lastBlock := bc.LastBlock()
		lastProof := lastBlock.Proof
		proof := bc.ProofOfWork(lastProof)
		bc.NewTransaction(
			"0",
			nodeIdentifier,
			1,
		)
		block, err := bc.NewBlock(proof, "")
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}

		res := &MineResponse{
			Message:      "Mined a new block",
			Index:        block.Index,
			Transactions: block.Transactions,
			Proof:        block.Proof,
			PreviousHash: block.PreviousHash,
		}

		return c.JSON(http.StatusOK, res)
	})
	e.GET("/chain", func(c echo.Context) error {
		res := &ChainResponse{
			Chain:  bc.Chain,
			Length: len(bc.Chain),
		}
		return c.JSON(http.StatusOK, res)
	})
	e.POST("/nodes/register", func(c echo.Context) error {
		b := new(RegisterBody)
		if err := c.Bind(b); err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		if len(b.Nodes) == 0 {
			return c.JSON(http.StatusBadRequest, "Invalid nodes")
		}
		for _, node := range b.Nodes {
			bc.RegisterNode(node)
		}
		res := &RegisterResponse{
			"Added a new node",
			bc.Nodes.ToSlice(),
		}
		return c.JSON(http.StatusCreated, res)
	})
	e.GET("/nodes/resolve", func(c echo.Context) error {
		replaced, err := bc.ResolveConflicts()
		if err != nil {
			return c.JSON(http.StatusBadRequest, err)
		}
		msg := "Chain was confirmed."
		if replaced {
			msg = "Chain was replaced."
		}
		res := &ResolveResponse{
			msg,
			bc.Chain,
		}
		return c.JSON(http.StatusOK, res)
	})
	e.Logger.Fatal(e.Start(":" + port))
}
