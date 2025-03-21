package crypto

// BlockchainAdapter defines the interface for chain-specific operations
type BlockchainAdapter interface {
	// Address operations
	GenerateWallet() (address string, privateKey string, err error)
	ValidateAddress(address string) (bool, error)
	
	// Transaction operations
	EstimateTransactionFee(params TransferParams) (*FeeEstimate, error)
	SendTransaction(privateKey string, params TransferParams) (txHash string, err error)
	GetTransactionStatus(txHash string) (TransactionStatus, int, error)
	
	// Balance operations
	GetBalance(address string, currency string) (string, error)
}