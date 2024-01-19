package main

import (
	"context"
	"crypto/ecdsa"
	"example/hello/api"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// Connect to infura node
    infuraNode := "https://sepolia.infura.io/v3/d7a41809b44c463b88bd585a3eaacc24"
     
	cl, err := ethclient.Dial(infuraNode)
	if err != nil {
		log.Fatalf("error dialing eth client: %v", err)
	}
	defer cl.Close()

	// Deploy Smart Contract
	contractAddress := deploySmartContract(cl)

	// Load Smart Contract
	contract := loadSmartContract(cl, contractAddress.Hex())
	txOptions := createTransaction(cl)

	// Use 'Store' operation of the contract
	tx1, err := contract.Store(txOptions, big.NewInt(10))
	if err != nil {
		log.Fatalf("unable to call store message function. %v", err)
	}
	waitForBlock(cl, tx1)
	log.Printf("Transferencia realizada satisfactoriamente")

	// Use 'Retrive' operation of the contract
    storedNumber, err := contract.Retrieve(&bind.CallOpts{})
	if err != nil {
		log.Printf("unable to call Symbol function. Err: %v", err)
	}
	log.Printf("Number stored in smart contract: %v", storedNumber)	
}

func deploySmartContract(cl *ethclient.Client) common.Address {
	txOptions := createTransaction(cl)

    address, txdc, _, err := api.DeployApi(txOptions, cl)
    if err != nil {
        log.Fatalf("unable to deploy smart contract. %v", err)
    }

    receipt := waitForBlock(cl, txdc)
    log.Printf("Smart Contract deployed successfully.")
	log.Printf("Deployed on Address: %s. Contract Address: %s. Gas Used: %v", address.Hex(), receipt.ContractAddress, receipt.GasUsed)

	return receipt.ContractAddress
}

func loadSmartContract(cl *ethclient.Client, address string) *api.Api {
	contractAddress:= common.HexToAddress(address)
	con, err := api.NewApi(contractAddress, cl)
	if err != nil {
		log.Fatalf("Unable to load smart contract")
	}
	log.Printf("Smart Contract loaded successfully.")
	return con
}

func createTransaction(cl *ethclient.Client) *bind.TransactOpts {
	// Use private key of metamask address
	privateKey := "eff3fc7fd51dde53de132e915235437a4a011dea29ac3c8097410e1c56bb4f54"
    key, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		log.Fatalf("Private key is not OK. %v", err)
	} 

	publicKey := key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatalf("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	gasPrice, err := cl.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatalf("error reading suggest gas price")
	}

	chainID, err := cl.ChainID(context.Background())
	if err != nil {
		log.Fatalf("unable to get chainID: %v", err)
	}
	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		log.Fatalf("unable to build new transactor: %v", err)
	}

	nonce, err := cl.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatalf("error reading next nonce")
	}
		
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(800000) // in units
	auth.GasPrice = gasPrice

	return auth
}

func waitForBlock(cl *ethclient.Client, tx *types.Transaction) *types.Receipt {
	for true {
		receipt, err := cl.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			log.Printf("Esperando 5seg")
			time.Sleep(5 * time.Second)
		} else {
			return receipt
		}
	}
	return nil
}