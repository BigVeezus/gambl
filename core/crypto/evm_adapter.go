// core/crypto/evm_adapter.go
package crypto

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// EVMAdapter implements BlockchainAdapter for Ethereum-compatible chains
type EVMAdapter struct {
	client  *ethclient.Client
	chainID *big.Int
	rpcURL  string
}

// NewEVMAdapter creates a new adapter for EVM-compatible chains
func NewEVMAdapter(rpcURL string, chainID int64) (*EVMAdapter, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &EVMAdapter{
		client:  client,
		chainID: big.NewInt(chainID),
		rpcURL:  rpcURL,
	}, nil
}

// GenerateWallet creates a new Ethereum wallet
func (a *EVMAdapter) GenerateWallet() (address string, privateKey string, err error) {
	// Generate private key
	key, err := crypto.GenerateKey()
	if err != nil {
		return "", "", err
	}

	// Get private key in hex format
	privateKeyBytes := crypto.FromECDSA(key)
	privateKeyHex := hexutil.Encode(privateKeyBytes)[2:] // Remove 0x prefix

	// Get public address
	publicKey := key.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", "", errors.New("error casting public key to ECDSA")
	}

	address = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return address, privateKeyHex, nil
}

// ValidateAddress checks if an address is a valid Ethereum address
func (a *EVMAdapter) ValidateAddress(address string) (bool, error) {
	// Simple regex check for Ethereum address format
	matched, err := regexp.MatchString("^0x[0-9a-fA-F]{40}$", address)
	if err != nil {
		return false, err
	}

	if !matched {
		return false, nil
	}

	// Convert to address type (this will check checksum)
	addr := common.HexToAddress(address)
	
	// If the input was a valid checksum address, HexToAddress will return the same
	// If not, it will be different
	checksumAddr := addr.Hex()
	
	// Check if the address was already checksummed
	if address == checksumAddr {
		return true, nil
	}
	
	// For non-checksummed addresses, we'll accept them but consider them valid
	return true, nil
}

// EstimateTransactionFee estimates the fee for a transaction
func (a *EVMAdapter) EstimateTransactionFee(params TransferParams) (*FeeEstimate, error) {
	ctx := context.Background()
	
	// Get current gas price from the network
	gasPrice, err := a.client.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}
	
	// For EIP-1559 chains, get base fee and priority fee
	header, err := a.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest block header: %w", err)
	}
	
	// Convert addresses to common.Address
	from := common.HexToAddress(params.FromAddress)
	to := common.HexToAddress(params.ToAddress)
	
	// Convert amount to wei
	amount, ok := new(big.Int).SetString(params.Amount, 10)
	if !ok {
		return nil, fmt.Errorf("invalid amount format")
	}
	
	// Estimate gas needed for the transaction
	gasLimit, err := a.client.EstimateGas(ctx, ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: amount,
		Data:  nil, // No data for simple transfers
	})
	if err != nil {
		// Default to standard ETH transfer gas (21000)
		gasLimit = 21000
	}
	
	// Calculate fee based on gas price and limit
	fee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	
	// If EIP-1559 is active (header.BaseFee is not nil)
	var baseFee, priorityFee, maxFee string
	if header.BaseFee != nil {
		// Get base fee from latest block
		baseFee = weiToEth(header.BaseFee).String()
		
		// Suggest priority fee (usually 1-2 Gwei)
		priorityFeeWei := big.NewInt(1500000000) // 1.5 Gwei
		priorityFee = weiToEth(priorityFeeWei).String()
		
		// Calculate max fee (base fee + priority fee + buffer)
		// Buffer is usually 2x base fee for safety
		buffer := new(big.Int).Mul(header.BaseFee, big.NewInt(2))
		maxFeeWei := new(big.Int).Add(
			new(big.Int).Add(header.BaseFee, priorityFeeWei),
			buffer,
		)
		maxFee = weiToEth(maxFeeWei).String()
		
		// Recalculate fee with EIP-1559 pricing
		fee = new(big.Int).Mul(
			maxFeeWei,
			big.NewInt(int64(gasLimit)),
		)
	} else {
		// Legacy gas pricing
		baseFee = "0"
		priorityFee = "0"
		maxFee = weiToEth(gasPrice).String()
	}
	
	// Convert wei to ETH
	feeEth := weiToEth(fee)
	
	return &FeeEstimate{
		EstimatedFee:     feeEth.String(),
		Currency:         params.Currency,
		BaseFee:          baseFee,
		PriorityFee:      priorityFee,
		MaxFee:           maxFee,
		EstimatedTimeMin: 1,
		EstimatedTimeMax: 5,
	}, nil
}

// SendTransaction sends a transaction on the Ethereum network
// SendTransaction sends a transaction on the Ethereum network
func (a *EVMAdapter) SendTransaction(privateKeyHex string, params TransferParams) (string, error) {
	ctx := context.Background()
	
	// Add "0x" prefix if not present
	if !strings.HasPrefix(privateKeyHex, "0x") {
		privateKeyHex = "0x" + privateKeyHex
	}
	
	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return "", fmt.Errorf("invalid private key: %w", err)
	}
	
	// Get public key and address
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", errors.New("error casting public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	
	// Verify from address matches
	if strings.ToLower(fromAddress.Hex()) != strings.ToLower(params.FromAddress) {
		return "", errors.New("from address does not match private key")
	}
	
	// Get nonce for the from address
	nonce, err := a.client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}
	
	// Convert amount to wei
	amount, ok := new(big.Int).SetString(params.Amount, 10)
	if !ok {
		return "", fmt.Errorf("invalid amount format")
	}
	
	// Convert to address
	toAddress := common.HexToAddress(params.ToAddress)
	
	// Create transaction
	var tx *types.Transaction
	
	// Get latest block header to check if EIP-1559 is active
	header, err := a.client.HeaderByNumber(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to get latest block header: %w", err)
	}
	
	// Set gas limit
	gasLimit := params.GasLimit
	if gasLimit == 0 {
		// Estimate gas if not provided
		gasLimit, err = a.client.EstimateGas(ctx, ethereum.CallMsg{
			From:  fromAddress,
			To:    &toAddress,
			Value: amount,
			Data:  nil, // No data for simple transfers
		})
		if err != nil {
			// Default to standard ETH transfer gas
			gasLimit = 21000
		}
	}
	
	// Check if EIP-1559 is active (header.BaseFee is not nil)
	if header.BaseFee != nil {
		// Create EIP-1559 transaction
		
		// Get suggested tip cap (priority fee)
		tipCap, err := a.client.SuggestGasTipCap(ctx)
		if err != nil {
			// Default to 1.5 Gwei if can't get suggested tip
			tipCap = big.NewInt(1500000000)
		}
		
		// Calculate fee cap (base fee + priority fee + buffer)
		// Buffer is usually 2x base fee for safety
		buffer := new(big.Int).Mul(header.BaseFee, big.NewInt(2))
		feeCap := new(big.Int).Add(
			new(big.Int).Add(header.BaseFee, tipCap),
			buffer,
		)
		
		// Create the transaction
		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   a.chainID,
			Nonce:     nonce,
			GasTipCap: tipCap,
			GasFeeCap: feeCap,
			Gas:       gasLimit,
			To:        &toAddress,
			Value:     amount,
			Data:      nil,
		})
	} else {
		// Create legacy transaction
		
		// Get gas price
		gasPrice, err := a.client.SuggestGasPrice(ctx)
		if err != nil {
			return "", fmt.Errorf("failed to get gas price: %w", err)
		}
		
		// If gas price was specified in the request, use that instead
		if params.GasPrice != "" {
			specifiedGasPrice, ok := new(big.Int).SetString(params.GasPrice, 10)
			if ok {
				gasPrice = specifiedGasPrice
			}
		}
		
		// Create the transaction
		tx = types.NewTx(&types.LegacyTx{
			Nonce:    nonce,
			GasPrice: gasPrice,
			Gas:      gasLimit,
			To:       &toAddress,
			Value:    amount,
			Data:     nil,
		})
	}
	
	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(a.chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}
	
	// Send the transaction
	err = a.client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}
	
	// Return the transaction hash
	return signedTx.Hash().Hex(), nil
}

// GetTransactionStatus checks the status of a transaction
func (a *EVMAdapter) GetTransactionStatus(txHash string) (TransactionStatus, int, error) {
	ctx := context.Background()
	
	// Convert hash string to common.Hash
	hash := common.HexToHash(txHash)
	
	// Get transaction receipt
	receipt, err := a.client.TransactionReceipt(ctx, hash)
	if err != nil {
		if errors.Is(err, ethereum.NotFound) {
			// Transaction is still pending (not mined yet)
			return TxStatusPending, 0, nil
		}
		return TxStatusPending, 0, err
	}
	
	// Get latest block number
	latestBlock, err := a.client.BlockNumber(ctx)
	if err != nil {
		return TxStatusPending, 0, err
	}
	
	// Calculate confirmations
	confirmations := int(latestBlock - receipt.BlockNumber.Uint64())
	
	// Check transaction status
	if receipt.Status == 1 {
		// Transaction succeeded
		return TxStatusConfirmed, confirmations, nil
	} else {
		// Transaction failed
		return TxStatusFailed, confirmations, nil
	}
}

// GetBalance gets the balance of an address
func (a *EVMAdapter) GetBalance(address string, currency string) (string, error) {
	ctx := context.Background()
	
	// Convert address string to common.Address
	addr := common.HexToAddress(address)
	
	// Get balance
	balance, err := a.client.BalanceAt(ctx, addr, nil) // nil = latest block
	if err != nil {
		return "", err
	}
	
	// Convert wei to ETH
	ethBalance := weiToEth(balance)
	
	// Return balance as string
	return ethBalance.String(), nil
}

// Helper function to convert wei to ETH
func weiToEth(wei *big.Int) *big.Float {
	// 1 ETH = 10^18 wei
	weiFloat := new(big.Float).SetInt(wei)
	ethFloat := new(big.Float).Quo(weiFloat, big.NewFloat(1e18))
	return ethFloat
}

// Helper function to convert ETH to wei
func ethToWei(eth string) (*big.Int, error) {
	// Parse ETH amount
	ethFloat, ok := new(big.Float).SetString(eth)
	if !ok {
		return nil, errors.New("invalid ETH amount")
	}
	
	// Convert to wei (multiply by 10^18)
	weiFloat := new(big.Float).Mul(ethFloat, big.NewFloat(1e18))
	
	// Convert to *big.Int
	wei := new(big.Int)
	weiFloat.Int(wei)
	
	return wei, nil
}