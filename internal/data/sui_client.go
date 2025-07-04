package data

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type SuiClient struct {
	rpcURL     string
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	address    string
	config     *Config
	httpClient *http.Client
}

type Config struct {
	PackageID  string `json:"package_id"`
	ModuleName string `json:"module_name"`
	PoolID     string `json:"pool_id"`
}

type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type TransactionBlockResponse struct {
	Digest  string                 `json:"digest"`
	Effects map[string]interface{} `json:"effects,omitempty"`
	Events  []SuiEvent             `json:"events,omitempty"`
}

type SuiEvent struct {
	Type       string          `json:"type"`
	ParsedJson json.RawMessage `json:"parsedJson"`
}

type MoveCallRequest struct {
	Signer          string        `json:"signer"`
	PackageObjectId string        `json:"packageObjectId"`
	Module          string        `json:"module"`
	Function        string        `json:"function"`
	TypeArguments   []string      `json:"typeArguments"`
	Arguments       []interface{} `json:"arguments"`
	Gas             string        `json:"gas,omitempty"`
	GasBudget       string        `json:"gasBudget"`
}

func NewSuiClient(rpcURL string, privateKeyStr string, cfg *Config) (*SuiClient, error) {
	var privateKeyBytes []byte
	var err error

	privateKeyBytes, err = hex.DecodeString(privateKeyStr)
	if err != nil {
		privateKeyBytes, err = base64.StdEncoding.DecodeString(privateKeyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode private key: %v", err)
		}
		// Sui CLI base64 keys include a 0x00 flag byte; use the 32-byte seed
		if len(privateKeyBytes) == 33 && privateKeyBytes[0] == 0x00 {
			privateKeyBytes = privateKeyBytes[1:33]
		}
	}

	// Remove 0x prefix if present
	if len(privateKeyStr) > 2 && privateKeyStr[:2] == "0x" {
		privateKeyStr = privateKeyStr[2:]
		privateKeyBytes, err = hex.DecodeString(privateKeyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex private key: %v", err)
		}
	}

	if len(privateKeyBytes) != 32 {
		return nil, fmt.Errorf("private key must be 32 bytes, got %d", len(privateKeyBytes))
	}

	privateKey := ed25519.NewKeyFromSeed(privateKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	// Generate Sui address
	address := fmt.Sprintf("0x%x", publicKey[:20])

	return &SuiClient{
		rpcURL:     rpcURL,
		privateKey: privateKey,
		publicKey:  publicKey,
		address:    address,
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (s *SuiClient) makeRPCCall(ctx context.Context, method string, params []interface{}) (*RPCResponse, error) {
	reqBody := RPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.rpcURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
		}
	}(resp.Body)

	var rpcResp RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC error [%d]: %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return &rpcResp, nil
}

func (s *SuiClient) GetCoins(ctx context.Context, coinType string) ([]map[string]interface{}, error) {
	params := []interface{}{
		s.address,
		coinType,
		nil,
		10,
	}

	resp, err := s.makeRPCCall(ctx, "suix_getCoins", params)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse coins response: %v", err)
	}

	return result.Data, nil
}

func (s *SuiClient) GetBalance(ctx context.Context) (uint64, error) {
	params := []interface{}{
		s.address,
		"0x2::sui::SUI",
	}

	resp, err := s.makeRPCCall(ctx, "suix_getBalance", params)
	if err != nil {
		return 0, err
	}

	var result struct {
		TotalBalance string `json:"totalBalance"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return 0, fmt.Errorf("failed to parse balance response: %v", err)
	}

	balance, err := strconv.ParseUint(result.TotalBalance, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance amount: %v", err)
	}

	return balance, nil
}

func (s *SuiClient) ExternalStake(requesterCoinID, accepterCoinID string, amount uint64, ctx context.Context) (string, error) {
	if requesterCoinID == "" || accepterCoinID == "" {
		return "", fmt.Errorf("both coin IDs are required")
	}
	if amount == 0 {
		return "", fmt.Errorf("stake amount must be greater than 0")
	}

	coins, err := s.GetCoins(ctx, "0x2::sui::SUI")
	if err != nil {
		return "", fmt.Errorf("failed to get coins for gas: %v", err)
	}

	var gasCoin string
	for _, coin := range coins {
		coinId := coin["coinObjectId"].(string)
		if coinId != requesterCoinID && coinId != accepterCoinID {
			gasCoin = coinId
			break
		}
	}

	if gasCoin == "" {
		return "", fmt.Errorf("no gas coin available")
	}

	// Build move call - matches blockchain function exactly
	//this call speaks to blockchain
	moveCallReq := MoveCallRequest{
		Signer:          s.address,
		PackageObjectId: s.config.PackageID,
		Module:          s.config.ModuleName,
		Function:        "external_stake",
		TypeArguments:   []string{"0x2::sui::SUI"},
		Arguments: []interface{}{
			s.config.PoolID,
			requesterCoinID,
			accepterCoinID,
			fmt.Sprintf("%d", amount),
		},
		Gas:       gasCoin,
		GasBudget: "10000000",
	}

	return s.executeTransaction(ctx, moveCallReq, "ExternalGameStaked")
}

func (s *SuiClient) ExternalPayWinner(requesterAddress, accepterAddress string, requesterScore, accepterScore, stakeAmount uint64, ctx context.Context) (string, error) {
	coins, err := s.GetCoins(ctx, "0x2::sui::SUI")
	if err != nil {
		return "", fmt.Errorf("failed to get coins: %v", err)
	}
	if len(coins) == 0 {
		return "", fmt.Errorf("no SUI coins available for gas")
	}
	moveCallReq := MoveCallRequest{
		Signer:          s.address,
		PackageObjectId: s.config.PackageID,
		Module:          s.config.ModuleName,
		Function:        "external_pay_winner",
		TypeArguments:   []string{},
		Arguments: []interface{}{
			s.config.PoolID,
			requesterAddress,
			accepterAddress,
			fmt.Sprintf("%d", requesterScore),
			fmt.Sprintf("%d", accepterScore),
			fmt.Sprintf("%d", stakeAmount),
		},
		Gas:       coins[0]["coinObjectId"].(string),
		GasBudget: "10000000", // 0.01 SUI
	}

	params := []interface{}{moveCallReq}
	resp, err := s.makeRPCCall(ctx, "unsafe_moveCall", params)
	if err != nil {
		return "", fmt.Errorf("failed to build transaction: %v", err)
	}

	var txBytes struct {
		TxBytes string `json:"txBytes"`
	}
	if err := json.Unmarshal(resp.Result, &txBytes); err != nil {
		return "", fmt.Errorf("failed to parse transaction bytes: %v", err)
	}

	signature, err := s.signTransaction(txBytes.TxBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	execParams := []interface{}{
		txBytes.TxBytes,
		[]string{signature},
		map[string]interface{}{
			"showEvents":  true,
			"showEffects": true,
		},
	}

	execResp, err := s.makeRPCCall(ctx, "sui_executeTransactionBlock", execParams)
	if err != nil {
		return "", fmt.Errorf("failed to execute transaction: %v", err)
	}

	var txResult TransactionBlockResponse
	if err := json.Unmarshal(execResp.Result, &txResult); err != nil {
		return "", fmt.Errorf("failed to parse execution result: %v", err)
	}

	eventType := fmt.Sprintf("%s::%s::ExternalGameCompleted", s.config.PackageID, s.config.ModuleName)
	for _, event := range txResult.Events {
		if event.Type == eventType {
			var winEvent struct {
				Requester      string `json:"requester"`
				Accepter       string `json:"accepter"`
				RequesterScore uint64 `json:"requester_score"`
				AccepterScore  uint64 `json:"accepter_score"`
				Winner         string `json:"winner"`
				PrizeAmount    uint64 `json:"prize_amount"`
				APIFee         uint64 `json:"api_fee"`
				EscrowFee      uint64 `json:"escrow_fee"`
				TotalStake     uint64 `json:"total_stake"`
				APICaller      string `json:"api_caller"`
				Timestamp      uint64 `json:"timestamp"`
			}
			if err := json.Unmarshal(event.ParsedJson, &winEvent); err == nil {
				fmt.Printf("✅ Win event: winner=%s, prize=%d, api_fee=%d, escrow_fee=%d, total_stake=%d\n",
					winEvent.Winner, winEvent.PrizeAmount, winEvent.APIFee, winEvent.EscrowFee, winEvent.TotalStake)
			}
		}
	}

	return txResult.Digest, nil
}

func (s *SuiClient) ExecuteTransactionBlock(ctx context.Context, txBytes []byte) (string, error) {
	txBytesStr := base64.StdEncoding.EncodeToString(txBytes)

	signature, err := s.signTransaction(txBytesStr)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	execParams := []interface{}{
		txBytesStr,
		[]string{signature},
		map[string]interface{}{
			"showEvents":  true,
			"showEffects": true,
		},
	}

	execResp, err := s.makeRPCCall(ctx, "sui_executeTransactionBlock", execParams)
	if err != nil {
		return "", fmt.Errorf("failed to execute transaction: %v", err)
	}

	var txResult TransactionBlockResponse
	if err := json.Unmarshal(execResp.Result, &txResult); err != nil {
		return "", fmt.Errorf("failed to parse execution result: %v", err)
	}

	return txResult.Digest, nil
}

func (s *SuiClient) GetTransactionBlock(ctx context.Context, digest string) (interface{}, error) {
	if digest == "" {
		return nil, fmt.Errorf("transaction digest is required")
	}

	params := []interface{}{
		digest,
		map[string]interface{}{
			"showInput":         true,
			"showRawInput":      false,
			"showEffects":       true,
			"showEvents":        true,
			"showObjectChanges": true,
		},
	}

	resp, err := s.makeRPCCall(ctx, "sui_getTransactionBlock", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction block: %v", err)
	}

	var result interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse transaction block response: %v", err)
	}

	return result, nil
}

func (s *SuiClient) BuildTransactionBlock(ctx context.Context, params interface{}) ([]byte, error) {
	if params == nil {
		return nil, fmt.Errorf("transaction parameters are required")
	}

	var rpcParams []interface{}

	switch p := params.(type) {
	case MoveCallRequest:
		rpcParams = []interface{}{p}
	case map[string]interface{}:
		moveCall := MoveCallRequest{
			Signer:          s.address,
			PackageObjectId: s.config.PackageID,
			Module:          s.config.ModuleName,
		}

		if function, ok := p["function"].(string); ok {
			moveCall.Function = function
		}
		if args, ok := p["arguments"].([]interface{}); ok {
			moveCall.Arguments = args
		}
		if typeArgs, ok := p["typeArguments"].([]string); ok {
			moveCall.TypeArguments = typeArgs
		}
		if gas, ok := p["gas"].(string); ok {
			moveCall.Gas = gas
		}
		if budget, ok := p["gasBudget"].(string); ok {
			moveCall.GasBudget = budget
		} else {
			moveCall.GasBudget = "10000000"
		}

		rpcParams = []interface{}{moveCall}
	default:
		rpcParams = []interface{}{params}
	}

	resp, err := s.makeRPCCall(ctx, "unsafe_moveCall", rpcParams)
	if err != nil {
		return nil, fmt.Errorf("failed to build transaction: %v", err)
	}

	var txBytes struct {
		TxBytes string `json:"txBytes"`
	}
	if err := json.Unmarshal(resp.Result, &txBytes); err != nil {
		return nil, fmt.Errorf("failed to parse transaction bytes: %v", err)
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(txBytes.TxBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to decode transaction bytes: %v", err)
	}

	return decodedBytes, nil
}

func (s *SuiClient) signTransaction(txBytes string) (string, error) {
	txBytesDecoded, err := base64.StdEncoding.DecodeString(txBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decode transaction bytes: %v", err)
	}

	signature := ed25519.Sign(s.privateKey, txBytesDecoded)

	// Encode signature (Sui format: flag + signature + public key)
	suiSignature := make([]byte, 1+len(signature)+len(s.publicKey))
	suiSignature[0] = 0x00 // Ed25519 flag
	copy(suiSignature[1:], signature)
	copy(suiSignature[1+len(signature):], s.publicKey)

	return base64.StdEncoding.EncodeToString(suiSignature), nil
}

func (s *SuiClient) splitCoin(ctx context.Context, coinId string, amount uint64) (string, error) {
	params := []interface{}{
		s.address,
		coinId,
		[]string{fmt.Sprintf("%d", amount)},
		nil,
		"10000000",
	}

	resp, err := s.makeRPCCall(ctx, "unsafe_splitCoin", params)
	if err != nil {
		return "", fmt.Errorf("failed to split coin: %v", err)
	}

	var txBytes struct {
		TxBytes string `json:"txBytes"`
	}
	if err := json.Unmarshal(resp.Result, &txBytes); err != nil {
		return "", fmt.Errorf("failed to parse split coin response: %v", err)
	}

	signature, err := s.signTransaction(txBytes.TxBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign split transaction: %v", err)
	}

	execParams := []interface{}{
		txBytes.TxBytes,
		[]string{signature},
		map[string]interface{}{
			"showEvents":  true,
			"showEffects": true,
		},
	}

	execResp, err := s.makeRPCCall(ctx, "sui_executeTransactionBlock", execParams)
	if err != nil {
		return "", fmt.Errorf("failed to execute split transaction: %v", err)
	}

	var txResult TransactionBlockResponse
	if err := json.Unmarshal(execResp.Result, &txResult); err != nil {
		return "", fmt.Errorf("failed to parse split execution result: %v", err)
	}

	return "new_coin_id", nil
}

func (s *SuiClient) GetSufficientCoins(ctx context.Context, requiredAmount uint64, count int) ([]string, error) {
	coins, err := s.GetCoins(ctx, "0x2::sui::SUI")
	if err != nil {
		return nil, fmt.Errorf("failed to get coins: %v", err)
	}

	var suitableCoins []string
	for _, coin := range coins {
		balance, err := strconv.ParseUint(coin["balance"].(string), 10, 64)
		if err != nil {
			continue
		}
		if balance >= requiredAmount {
			suitableCoins = append(suitableCoins, coin["coinObjectId"].(string))
			if len(suitableCoins) >= count {
				break
			}
		}
	}

	if len(suitableCoins) < count {
		return nil, fmt.Errorf("insufficient coins: need %d coins with %d balance each, found %d",
			count, requiredAmount, len(suitableCoins))
	}

	return suitableCoins, nil
}

func (s *SuiClient) MergeCoins(ctx context.Context, primaryCoin string, coinToMerge string) (string, error) {
	coins, err := s.GetCoins(ctx, "0x2::sui::SUI")
	if err != nil {
		return "", fmt.Errorf("failed to get coins for gas: %v", err)
	}

	var gasCoin string
	for _, coin := range coins {
		coinId := coin["coinObjectId"].(string)
		if coinId != primaryCoin && coinId != coinToMerge {
			gasCoin = coinId
			break
		}
	}

	if gasCoin == "" {
		return "", fmt.Errorf("no gas coin available for merge operation")
	}

	params := []interface{}{
		s.address,
		primaryCoin,
		coinToMerge,
		gasCoin,
		"10000000",
	}
	resp, err := s.makeRPCCall(ctx, "unsafe_mergeCoins", params)
	if err != nil {
		return "", fmt.Errorf("failed to merge coins: %v", err)
	}

	var txBytes struct {
		TxBytes string `json:"txBytes"`
	}
	if err := json.Unmarshal(resp.Result, &txBytes); err != nil {
		return "", fmt.Errorf("failed to parse merge response: %v", err)
	}

	signature, err := s.signTransaction(txBytes.TxBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign merge transaction: %v", err)
	}

	execParams := []interface{}{
		txBytes.TxBytes,
		[]string{signature},
		map[string]interface{}{
			"showEvents":  true,
			"showEffects": true,
		},
	}

	execResp, err := s.makeRPCCall(ctx, "sui_executeTransactionBlock", execParams)
	if err != nil {
		return "", fmt.Errorf("failed to execute merge transaction: %v", err)
	}

	var txResult TransactionBlockResponse
	if err := json.Unmarshal(execResp.Result, &txResult); err != nil {
		return "", fmt.Errorf("failed to parse merge execution result: %v", err)
	}

	return txResult.Digest, nil
}

func (s *SuiClient) ExternalStakeEnhanced(requesterAddress, accepterAddress string, amount uint64, ctx context.Context) (string, error) {
	if requesterAddress == "" || accepterAddress == "" {
		return "", fmt.Errorf("requester and accepter addresses are required")
	}
	if amount == 0 {
		return "", fmt.Errorf("stake amount must be greater than 0")
	}

	coins, err := s.GetSufficientCoins(ctx, amount, 2)
	if err != nil {
		return "", fmt.Errorf("failed to get sufficient coins for staking: %v", err)
	}

	allCoins, err := s.GetCoins(ctx, "0x2::sui::SUI")
	if err != nil {
		return "", fmt.Errorf("failed to get coins for gas: %v", err)
	}

	var gasCoin string
	for _, coin := range allCoins {
		coinId := coin["coinObjectId"].(string)
		if coinId != coins[0] && coinId != coins[1] {
			gasCoin = coinId
			break
		}
	}

	if gasCoin == "" {
		return "", fmt.Errorf("no gas coin available")
	}

	moveCallReq := MoveCallRequest{
		Signer:          s.address,
		PackageObjectId: s.config.PackageID,
		Module:          s.config.ModuleName,
		Function:        "external_stake",
		TypeArguments:   []string{"0x2::sui::SUI"},
		Arguments: []interface{}{
			s.config.PoolID,
			coins[0],
			coins[1],
			fmt.Sprintf("%d", amount),
		},
		Gas: gasCoin,
		// gas budget increased because of complexity of the transaction
		GasBudget: "20000000",
	}

	return s.executeTransaction(ctx, moveCallReq, "ExternalGameStaked")
}

func (s *SuiClient) executeTransaction(ctx context.Context, moveCallReq MoveCallRequest, expectedEventSuffix string) (string, error) {
	params := []interface{}{moveCallReq}
	resp, err := s.makeRPCCall(ctx, "unsafe_moveCall", params)
	if err != nil {
		return "", fmt.Errorf("failed to build transaction: %v", err)
	}

	var txBytes struct {
		TxBytes string `json:"txBytes"`
	}
	if err := json.Unmarshal(resp.Result, &txBytes); err != nil {
		return "", fmt.Errorf("failed to parse transaction bytes: %v", err)
	}

	signature, err := s.signTransaction(txBytes.TxBytes)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	execParams := []interface{}{
		txBytes.TxBytes,
		[]string{signature},
		map[string]interface{}{
			"showEvents":  true,
			"showEffects": true,
		},
	}

	execResp, err := s.makeRPCCall(ctx, "sui_executeTransactionBlock", execParams)
	if err != nil {
		return "", fmt.Errorf("failed to execute transaction: %v", err)
	}
	var txResult TransactionBlockResponse
	if err := json.Unmarshal(execResp.Result, &txResult); err != nil {
		return "", fmt.Errorf("failed to parse execution result: %v", err)
	}

	if effects, ok := txResult.Effects["status"]; ok {
		if status, ok := effects.(map[string]interface{}); ok {
			if status["status"] != "success" {
				return "", fmt.Errorf("transaction failed with status: %v", status)
			}
		}
	}

	eventType := fmt.Sprintf("%s::%s::%s", s.config.PackageID, s.config.ModuleName, expectedEventSuffix)
	eventFound := false
	for _, event := range txResult.Events {
		if event.Type == eventType {
			eventFound = true
			fmt.Printf("✅ Event emitted: %s\n", string(event.ParsedJson))
			break
		}
	}

	if !eventFound {
		fmt.Printf("⚠️  Expected event %s not found in transaction events\n", eventType)
	}

	return txResult.Digest, nil
}

func (s *SuiClient) GetAddress() string {
	return s.address
}

func (s *SuiClient) GetStakePool(poolID string) (map[string]interface{}, error) {
	if poolID == "" {
		return nil, fmt.Errorf("pool ID is required")
	}

	params := []interface{}{
		poolID,
		map[string]bool{
			"showContent": true,
			"showType":    true,
		},
	}

	resp, err := s.makeRPCCall(context.Background(), "sui_getObject", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get stake pool: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse stake pool response: %v", err)
	}

	if data, ok := result["data"]; ok {
		if dataMap, ok := data.(map[string]interface{}); ok {
			if dataMap["objectId"] == nil {
				return nil, fmt.Errorf("stake pool not found")
			}
		}
	}

	return result, nil
}

func (s *SuiClient) ValidatePoolConfig(ctx context.Context) error {
	if s.config.PoolID == "" {
		return fmt.Errorf("pool ID not configured")
	}
	if s.config.PackageID == "" {
		return fmt.Errorf("package ID not configured")
	}
	if s.config.ModuleName == "" {
		return fmt.Errorf("module name not configured")
	}
	_, err := s.GetStakePool(s.config.PoolID)
	if err != nil {
		return fmt.Errorf("invalid pool configuration: %v", err)
	}

	return nil
}

func (s *SuiClient) GetCoinsByType(ctx context.Context, coinType string, cursor *string, limit int) ([]map[string]interface{}, *string, error) {
	params := []interface{}{
		s.address,
		coinType,
		cursor,
		limit,
	}

	resp, err := s.makeRPCCall(ctx, "suix_getCoins", params)
	if err != nil {
		return nil, nil, err
	}

	var result struct {
		Data        []map[string]interface{} `json:"data"`
		NextCursor  *string                  `json:"nextCursor"`
		HasNextPage bool                     `json:"hasNextPage"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, nil, fmt.Errorf("failed to parse coins response: %v", err)
	}

	var nextCursor *string
	if result.HasNextPage {
		nextCursor = result.NextCursor
	}

	return result.Data, nextCursor, nil
}

func (s *SuiClient) GetAllCoins(ctx context.Context, coinType string) ([]map[string]interface{}, error) {
	var allCoins []map[string]interface{}
	var cursor *string

	for {
		coins, nextCursor, err := s.GetCoinsByType(ctx, coinType, cursor, 50)
		if err != nil {
			return nil, err
		}

		allCoins = append(allCoins, coins...)

		if nextCursor == nil {
			break
		}
		cursor = nextCursor
	}

	return allCoins, nil
}

func (s *SuiClient) GetTotalBalance(ctx context.Context, coinType string) (uint64, error) {
	coins, err := s.GetAllCoins(ctx, coinType)
	if err != nil {
		return 0, err
	}

	var totalBalance uint64
	for _, coin := range coins {
		if balanceStr, ok := coin["balance"].(string); ok {
			balance, err := strconv.ParseUint(balanceStr, 10, 64)
			if err != nil {
				continue
			}
			totalBalance += balance
		}
	}

	return totalBalance, nil
}

func (s *SuiClient) WaitForTransaction(ctx context.Context, digest string, maxWaitTime time.Duration) (interface{}, error) {
	if digest == "" {
		return nil, fmt.Errorf("transaction digest is required")
	}

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return nil, fmt.Errorf("timeout waiting for transaction %s", digest)
		case <-ticker.C:
			tx, err := s.GetTransactionBlock(ctx, digest)
			if err == nil {
				return tx, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (s *SuiClient) EstimateGas(ctx context.Context, txBytes []byte) (uint64, error) {
	txBytesStr := base64.StdEncoding.EncodeToString(txBytes)

	params := []interface{}{
		txBytesStr,
	}

	resp, err := s.makeRPCCall(ctx, "sui_dryRunTransactionBlock", params)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %v", err)
	}

	var result struct {
		Effects struct {
			GasUsed struct {
				ComputationCost uint64 `json:"computationCost,string"`
				StorageCost     uint64 `json:"storageCost,string"`
				StorageRebate   uint64 `json:"storageRebate,string"`
			} `json:"gasUsed"`
		} `json:"effects"`
	}

	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return 0, fmt.Errorf("failed to parse gas estimation: %v", err)
	}

	totalGas := result.Effects.GasUsed.ComputationCost + result.Effects.GasUsed.StorageCost - result.Effects.GasUsed.StorageRebate
	return totalGas, nil
}

func (s *SuiClient) GetObjectsOwnedByAddress(ctx context.Context, objectType *string) ([]map[string]interface{}, error) {
	params := []interface{}{
		s.address,
	}

	if objectType != nil {
		params = append(params, map[string]interface{}{
			"filter": map[string]string{
				"StructType": *objectType,
			},
		})
	}

	resp, err := s.makeRPCCall(ctx, "suix_getOwnedObjects", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get owned objects: %v", err)
	}

	var result struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse owned objects response: %v", err)
	}

	return result.Data, nil
}

func (s *SuiClient) HealthCheck(ctx context.Context) error {
	resp, err := s.makeRPCCall(ctx, "sui_getLatestCheckpointSequenceNumber", []interface{}{})
	if err != nil {
		return fmt.Errorf("health check failed: %v", err)
	}

	var checkpoint string
	if err := json.Unmarshal(resp.Result, &checkpoint); err != nil {
		return fmt.Errorf("health check failed to parse response: %v", err)
	}
	if checkpoint == "" {
		return fmt.Errorf("health check failed: empty checkpoint sequence number")
	}

	return nil
}

func (s *SuiClient) GetNetworkInfo(ctx context.Context) (map[string]interface{}, error) {
	resp, err := s.makeRPCCall(ctx, "rpc.discover", []interface{}{})
	if err != nil {
		return nil, fmt.Errorf("failed to get network info: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse network info: %v", err)
	}

	return result, nil
}
