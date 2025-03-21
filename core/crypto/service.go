// core/crypto/service.go
package crypto

import (
	"context"
	// "errors"
	"log"
	"time"

	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

// cryptoService implements CryptoService interface
type cryptoService struct {
	//TODO: add payinchannelCollection. for deposits and transaction history collection
	payoutChannelCollection *mongo.Collection
	// txCollection *mongo.Collection
	adapters     map[ChainType]BlockchainAdapter
	encrypter    CryptoEncrypter
	logger       *log.Logger
}

// NewCryptoService creates a new crypto service instance
func NewCryptoService(
	payoutChannelCollection *mongo.Collection,
	// txCollection *mongo.Collection,
	adapters map[ChainType]BlockchainAdapter,
	encrypter CryptoEncrypter,
	logger *log.Logger,
) CryptoService {
	return &cryptoService{
		payoutChannelCollection: payoutChannelCollection,
		// txCollection: txCollection,
		adapters:     adapters,
		encrypter:    encrypter,
		logger:       logger,
	}
}

// GenerateAddress generates a new wallet address for a user
func (s *cryptoService) GenerateAddress(ctx context.Context, userID string, chainType ChainType) (*PayinDetails, error) {
	// Check if user already has an address for this chain
	// existing, err := s.GetUserAddress(ctx, userID, chainType)
	// if err == nil {
	// 	// User already has an address for this chain
	// 	return existing, nil
	// } else if !errors.Is(err, mongo.ErrNoDocuments) {
	// 	// Some other error occurred
	// 	return nil, err
	// }

	// Get the appropriate adapter for the chain
	adapter, ok := s.adapters[chainType]
	if !ok {
		return nil, ErrUnsupportedChain
	}

	// Generate a new wallet
	address, privateKey, err := adapter.GenerateWallet()
	if err != nil {
		s.logger.Printf("Error generating wallet: %v", err)
		return nil, ErrAddressGenerationFailed
	}

	// Encrypt the private key
	encryptedKey, err := s.encrypter.Encrypt(privateKey)
	if err != nil {
		s.logger.Printf("Error encrypting private key: %v", err)
		return nil, ErrEncryptionFailed
	}

	// Create wallet address record
	now := time.Now()
	payinDetails := &PayinDetails{
		ID:           primitive.NewObjectID().Hex(),
		UserID:       userID,
		Address:      address,
		ChainType:    chainType,
		EncryptedKey: encryptedKey,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// // Store in database
	// _, err = s.collection.InsertOne(ctx, payinDetails)
	// if err != nil {
	// 	s.logger.Printf("Error storing wallet address: %v", err)
	// 	return nil, err
	// }

	return payinDetails, nil
}

// GetUserAddress retrieves a user's payin details for a specific chain
// func (s *cryptoService) GetUserAddress(ctx context.Context, userID string, chainType ChainType) (*payinDetails, error) {
// 	var payinDetails payinDetails
// 	err := s.collection.FindOne(ctx, bson.M{
// 		"user_id":    userID,
// 		"chain_type": chainType,
// 	}).Decode(&payinDetails)

// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			return nil, mongo.ErrNoDocuments
// 		}
// 		s.logger.Printf("Error retrieving wallet address: %v", err)
// 		return nil, err
// 	}

// 	return &payinDetails, nil
// }

// ValidateAddress validates a wallet address for a specific chain
func (s *cryptoService) ValidateAddress(ctx context.Context, address string, chainType ChainType) (bool, error) {
	adapter, ok := s.adapters[chainType]
	if !ok {
		return false, ErrUnsupportedChain
	}

	return adapter.ValidateAddress(address)
}

// EstimateFee estimates the fee for a transaction
func (s *cryptoService) EstimateFee(ctx context.Context, params TransferParams) (*FeeEstimate, error) {
	adapter, ok := s.adapters[params.ChainType]
	if !ok {
		return nil, ErrUnsupportedChain
	}

	return adapter.EstimateTransactionFee(params)
}

// Transfer sends a crypto transaction
// func (s *cryptoService) Transfer(ctx context.Context, userID string, params TransferParams) (*Transaction, error) {
// 	// Get user's wallet for this chain
// 	wallet, err := s.GetUserAddress(ctx, userID, params.ChainType)
// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			return nil, errors.New("user does not have a wallet for this chain")
// 		}
// 		return nil, err
// 	}

// 	// Ensure from address matches the user's wallet
// 	if wallet.Address != params.FromAddress {
// 		return nil, errors.New("from address does not match user's wallet")
// 	}

// 	// Decrypt private key
// 	privateKey, err := s.encrypter.Decrypt(wallet.EncryptedKey)
// 	if err != nil {
// 		s.logger.Printf("Error decrypting private key: %v", err)
// 		return nil, ErrDecryptionFailed
// 	}

// 	// Get the appropriate adapter
// 	adapter, ok := s.adapters[params.ChainType]
// 	if !ok {
// 		return nil, ErrUnsupportedChain
// 	}

// 	// Create a transaction record
// 	now := time.Now()
// 	tx := &Transaction{
// 		ID:          primitive.NewObjectID().Hex(),
// 		UserID:      userID,
// 		FromAddress: params.FromAddress,
// 		ToAddress:   params.ToAddress,
// 		Amount:      params.Amount,
// 		Currency:    params.Currency,
// 		ChainType:   params.ChainType,
// 		Status:      TxStatusPending,
// 		CreatedAt:   now,
// 		UpdatedAt:   now,
// 		MetaData:    map[string]interface{}{},
// 	}

// 	// Store the transaction first with pending status
// 	_, err = s.txCollection.InsertOne(ctx, tx)
// 	if err != nil {
// 		s.logger.Printf("Error storing transaction: %v", err)
// 		return nil, err
// 	}

// 	// Send the transaction
// 	txHash, err := adapter.SendTransaction(privateKey, params)
// 	if err != nil {
// 		// Update transaction status to failed
// 		_, updateErr := s.txCollection.UpdateOne(
// 			ctx,
// 			bson.M{"_id": tx.ID},
// 			bson.M{
// 				"$set": bson.M{
// 					"status":          TxStatusFailed,
// 					"updated_at":      time.Now(),
// 					"metadata.error":  err.Error(),
// 				},
// 			},
// 		)
// 		if updateErr != nil {
// 			s.logger.Printf("Error updating failed transaction: %v", updateErr)
// 		}
// 		return nil, err
// 	}

// 	// Update transaction with hash
// 	_, err = s.txCollection.UpdateOne(
// 		ctx,
// 		bson.M{"_id": tx.ID},
// 		bson.M{
// 			"$set": bson.M{
// 				"tx_hash":    txHash,
// 				"updated_at": time.Now(),
// 			},
// 		},
// 	)
// 	if err != nil {
// 		s.logger.Printf("Error updating transaction with hash: %v", err)
// 	}

// 	tx.TxHash = txHash
// 	return tx, nil
// }

// GetTransactionStatus gets the status of a transaction
// func (s *cryptoService) GetTransactionStatus(ctx context.Context, txID string) (*Transaction, error) {
// 	// Retrieve transaction from database
// 	var tx Transaction
// 	err := s.txCollection.FindOne(ctx, bson.M{"_id": txID}).Decode(&tx)
// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			return nil, errors.New("transaction not found")
// 		}
// 		return nil, err
// 	}

// 	// If transaction is already in a final state, just return it
// 	if tx.Status == TxStatusConfirmed || tx.Status == TxStatusFailed {
// 		return &tx, nil
// 	}

// 	// If transaction is pending and has a hash, check status on-chain
// 	if tx.Status == TxStatusPending && tx.TxHash != "" {
// 		adapter, ok := s.adapters[tx.ChainType]
// 		if !ok {
// 			return nil, ErrUnsupportedChain
// 		}

// 		// Check transaction status on-chain
// 		status, confirmations, err := adapter.GetTransactionStatus(tx.TxHash)
// 		if err != nil {
// 			s.logger.Printf("Error checking transaction status: %v", err)
// 			return &tx, nil // Return current status on error
// 		}

// 		// Update transaction in database if status changed
// 		if status != tx.Status || confirmations != tx.Confirmations {
// 			update := bson.M{
// 				"status":        status,
// 				"confirmations": confirmations,
// 				"updated_at":    time.Now(),
// 			}

// 			// If the transaction is now confirmed, set completed timestamp
// 			if status == TxStatusConfirmed && tx.Status != TxStatusConfirmed {
// 				now := time.Now()
// 				update["completed_at"] = now
// 			}

// 			_, err := s.txCollection.UpdateOne(
// 				ctx,
// 				bson.M{"_id": tx.ID},
// 				bson.M{"$set": update},
// 			)
// 			if err != nil {
// 				s.logger.Printf("Error updating transaction status: %v", err)
// 			}

// 			// Update local tx object
// 			tx.Status = status
// 			tx.Confirmations = confirmations
// 			tx.UpdatedAt = time.Now()
// 			if status == TxStatusConfirmed && tx.CompletedAt == nil {
// 				now := time.Now()
// 				tx.CompletedAt = &now
// 			}
// 		}
// 	}

// 	return &tx, nil
// }

// GetTransactionsByUser gets all transactions for a user
// func (s *cryptoService) GetTransactionsByUser(ctx context.Context, userID string, limit, offset int) ([]Transaction, error) {
// 	opts := options.Find()
	
// 	// Add sorting, newest first
// 	opts.SetSort(bson.M{"created_at": -1})
	
// 	// Add pagination
// 	if limit > 0 {
// 		opts.SetLimit(int64(limit))
// 	}
// 	if offset > 0 {
// 		opts.SetSkip(int64(offset))
// 	}
	
// 	// Execute query
// 	cursor, err := s.txCollection.Find(
// 		ctx,
// 		bson.M{"user_id": userID},
// 		opts,
// 	)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer cursor.Close(ctx)
	
// 	// Decode results
// 	var transactions []Transaction
// 	if err := cursor.All(ctx, &transactions); err != nil {
// 		return nil, err
// 	}
	
// 	return transactions, nil
// }

// GetBalance gets the balance of an address
func (s *cryptoService) GetBalance(ctx context.Context, address string, currency string, chainType ChainType) (*BalanceInfo, error) {
	adapter, ok := s.adapters[chainType]
	if !ok {
		return nil, ErrUnsupportedChain
	}

	// Validate address format
	valid, err := adapter.ValidateAddress(address)
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, ErrInvalidAddress
	}

	// Get balance from blockchain
	balance, err := adapter.GetBalance(address, currency)
	if err != nil {
		return nil, err
	}

	// Return balance info
	return &BalanceInfo{
		Address:   address,
		Currency:  currency,
		Balance:   balance,
		Available: balance, // Default to same as balance
		ChainType: chainType,
		Timestamp: time.Now(),
	}, nil
}

// GetUserBalance gets the balance for a user's address
// func (s *cryptoService) GetUserBalance(ctx context.Context, userID string, currency string, chainType ChainType) (*BalanceInfo, error) {
// 	// Get user's wallet address
// 	wallet, err := s.GetUserAddress(ctx, userID, chainType)
// 	if err != nil {
// 		if errors.Is(err, mongo.ErrNoDocuments) {
// 			return nil, errors.New("user does not have a wallet for this chain")
// 		}
// 		return nil, err
// 	}
	
// 	// Get balance for the address
// 	return s.GetBalance(ctx, wallet.Address, currency, chainType)
// }