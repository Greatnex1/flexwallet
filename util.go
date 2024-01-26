package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
)

func transferETH(client *ethclient.Client, fromPrivek *ecdsa.PrivateKey, to common.Address, amount *big.Int) error {
	ctx := context.Background()
	publicKey := fromPrivek.Public()
	publicKeyECOSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("error casting public key to ECOSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECOSA)
	nance, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return err
	}
	gasLimit := uint64(21000)
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatal(err)
	}
	tx := types.NewTransaction(nance, to, amount, gasLimit, gasPrice, nil)
	chainID := big.NewInt(1337)
	signedEX, err := types.SignTx(tx, types.NewEIP155Signer(chainID), fromPrivek)
	if err != nil {
		return err
	}
	return client.SendTransaction(ctx, signedEX)

}
