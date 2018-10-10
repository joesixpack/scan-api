/**
*  @file
*  @copyright defined in scan-api/LICENSE
 */

package rpc

import (
	"fmt"
	"math/big"
)

// CurrentBlock returns the current block info.
func (rpc *SeeleRPC) CurrentBlock() (currentBlock *CurrentBlock, err error) {
	request := GetBlockByHeightRequest{
		Height: -1,
		FullTx: true,
	}
	rpcOutputBlock := make(map[string]interface{})
	if err := rpc.call("seele.GetBlockByHeight", request, &rpcOutputBlock); err != nil {
		return nil, err
	}

	timestamp := int64(rpcOutputBlock["timestamp"].(float64))
	difficulty := int64(rpcOutputBlock["difficulty"].(float64))
	height := uint64(rpcOutputBlock["height"].(float64))

	currentBlock = &CurrentBlock{
		HeadHash:  rpcOutputBlock["hash"].(string),
		Height:    height,
		Timestamp: big.NewInt(timestamp),
		Difficult: big.NewInt(difficulty),
		Creator:   rpcOutputBlock["creator"].(string),
		TxCount:   len(rpcOutputBlock["transactions"].([]interface{})),
	}
	return currentBlock, err
}

//GetBlockByHeight get block and transaction data from seele node
func (rpc *SeeleRPC) GetBlockByHeight(h uint64, fullTx bool) (block *BlockInfo, err error) {
	request := GetBlockByHeightRequest{
		Height: int64(h),
		FullTx: fullTx,
	}
	rpcOutputBlock := make(map[string]interface{})
	if err := rpc.call("seele.GetBlockByHeight", request, &rpcOutputBlock); err != nil {
		return nil, err
	}

	height := uint64(rpcOutputBlock["height"].(float64))
	hash := rpcOutputBlock["hash"].(string)
	parentHash := rpcOutputBlock["parentHash"].(string)
	nonce := uint64(rpcOutputBlock["nonce"].(float64))
	stateHash := rpcOutputBlock["stateHash"].(string)
	txHash := rpcOutputBlock["txHash"].(string)
	creator := rpcOutputBlock["creator"].(string)
	timestamp := int64(rpcOutputBlock["timestamp"].(float64))
	difficulty := int64(rpcOutputBlock["difficulty"].(float64))
	totalDifficulty := int64(rpcOutputBlock["totalDifficulty"].(float64))

	var Txs []Transaction
	if fullTx {
		var rpcTxs []interface{}
		rpcTxs = rpcOutputBlock["transactions"].([]interface{})
		for i := 0; i < len(rpcTxs); i++ {
			var tx Transaction
			rpcTx := rpcTxs[i].(map[string]interface{})
			tx.Hash = rpcTx["hash"].(string)
			tx.From = rpcTx["from"].(string)
			tx.To = rpcTx["to"].(string)
			amount := int64(rpcTx["amount"].(float64))
			tx.Amount = big.NewInt(amount)
			tx.AccountNonce = uint64(rpcTx["accountNonce"].(float64))
			tx.Payload = rpcTx["payload"].(string)
			tx.Timestamp = uint64(rpcTx["timestamp"].(float64))
			tx.Fee = int64(rpcTx["fee"].(float64))
			Txs = append(Txs, tx)
		}
	}

	block = &BlockInfo{
		Height:          height,
		Hash:            hash,
		ParentHash:      parentHash,
		Nonce:           nonce,
		StateHash:       stateHash,
		TxHash:          txHash,
		Creator:         creator,
		Timestamp:       big.NewInt(timestamp),
		Difficulty:      big.NewInt(difficulty),
		TotalDifficulty: big.NewInt(totalDifficulty),
		Txs:             Txs,
	}
	return block, err
}

// GetPeersInfo get peers info from connected seele node
func (rpc *SeeleRPC) GetPeersInfo() (result []PeerInfo, err error) {
	rpcPeerInfos := make([]map[string]interface{}, 0)
	if err := rpc.call("network_getPeersInfo", nil, &rpcPeerInfos); err != nil {
		return nil, err
	}

	// result data struct:
	// []map[
	//   id:0x0ea2a45ab5a909c309439b0e004c61b7b2a3e831
	//   caps:[lightSeele_1/1 lightSeele_2/1 seele/1]
	//   network:
	//     map[
	//       localAddress:127.0.0.1:8057
	//       remoteAddress:127.0.0.1:54337
	//     ]
	//   protocols:
	//     map[
	//       lightSeele_2:handshake
	//       seele:
	//         map[
	//           version:1
	//           difficulty:7.926036971e+09
	//           head:0000017b5835582b259848c6b0e21d35d90408205c1a41e0aeebe6a67797b8a8
	//         ]
	//       lightSeele_1:handshake
	//     ]
	//   shard:2
	// ]
	var peerInfos []PeerInfo
	for _, rpcPeerInfo := range rpcPeerInfos {
		id := rpcPeerInfo["id"].(string)
		rpcCaps := rpcPeerInfo["caps"].([]interface{})
		var caps []string
		for j := 0; j < len(rpcCaps); j++ {
			capString := rpcCaps[j].(string)
			caps = append(caps, capString)
		}
		rpcPeerNetWork := rpcPeerInfo["network"].(map[string]interface{})
		localAddress := rpcPeerNetWork["localAddress"].(string)
		remoteAddress := rpcPeerNetWork["remoteAddress"].(string)
		shardNumber := int(rpcPeerInfo["shard"].(float64))

		peerInfo := PeerInfo{
			ID:            id,
			Caps:          caps,
			LocalAddress:  localAddress,
			RemoteAddress: remoteAddress,
			ShardNumber:   shardNumber,
		}

		peerInfos = append(peerInfos, peerInfo)
	}

	return peerInfos, nil
}

// GetBalance get the balance of the account
func (rpc *SeeleRPC) GetBalance(address string) (int64, error) {
	result := make(map[string]interface{})
	if err := rpc.call("seele_getBalance", &address, &result); err != nil {
		return 0, err
	}

	// result data struct:
	// map[
	//   Balance:1.9975499e+12
	//   Account:0x4c10f2cd2159bb432094e3be7e17904c2b4aeb21
	// ]
	account := result["Account"].(string)
	if account != address {
		return 0, fmt.Errorf("expected balance '%s', actually '%s'", address, result)
	}
	balance := int64(result["Balance"].(float64))
	return balance, nil
}

// GetReceiptByTxHash get the receipt by tx hash
func (rpc *SeeleRPC) GetReceiptByTxHash(txhash string) (*Receipt, error) {
	rpcOutputReceipt := make(map[string]interface{})
	if err := rpc.call("txpool_getReceiptByTxHash", &txhash, &rpcOutputReceipt); err != nil {
		return nil, err
	}

	// result data struct:
	// map[
	//   poststate:0x95645120bcdc5f07dc3b8f30f0f3d4069d3374cf0167575f8be474d6c3ad7038
	//   result:0x
	//   totalFee:1
	//   txhash:0x02c240f019adc8b267b82026aef6b677c67867624e2acc1418149e7f8083ba0e
	//   usedGas:0
	//   contract:0x
	//   failed:false
	// ]
	result := rpcOutputReceipt["result"].(string)
	postState := rpcOutputReceipt["poststate"].(string)
	txHash := rpcOutputReceipt["txhash"].(string)
	contractAddress := rpcOutputReceipt["contract"].(string)
	failed := rpcOutputReceipt["failed"].(bool)
	totalFee := int64(rpcOutputReceipt["totalFee"].(float64))
	usedGas := int64(rpcOutputReceipt["usedGas"].(float64))

	receipt := Receipt{
		Result:          result,
		PostState:       postState,
		TxHash:          txHash,
		ContractAddress: contractAddress,
		Failed:          failed,
		TotalFee:        big.NewInt(totalFee),
		UsedGas:         big.NewInt(usedGas),
	}
	return &receipt, nil
}

// GetPendingTransactions get pending transactions on seele node
func (rpc *SeeleRPC) GetPendingTransactions() ([]Transaction, error) {
	rpcOutputTxs := make([]map[string]interface{}, 0)
	if err := rpc.call("debug_getPendingTransactions", nil, &rpcOutputTxs); err != nil {
		return nil, err
	}

	// result data struct:
	// []map[
	//   from:0x4c10f2cd2159bb432094e3be7e17904c2b4aeb21
	//   hash:0x6524d63226943b2c0cafca124983faa2c64dc2bacf27aab22f6b3ebc67404c39
	//   payload:
	//   timestamp:0
	//   to:0x0ea2a45ab5a909c309439b0e004c61b7b2a3e831
	//   accountNonce:14
	//   amount:10000
	//   fee:1
	// ]
	var Txs []Transaction
	for _, rpcTx := range rpcOutputTxs {
		var tx Transaction
		tx.Hash = rpcTx["hash"].(string)
		tx.From = rpcTx["from"].(string)
		tx.To = rpcTx["to"].(string)
		amount := int64(rpcTx["amount"].(float64))
		tx.Amount = big.NewInt(amount)
		tx.AccountNonce = uint64(rpcTx["accountNonce"].(float64))
		tx.Payload = rpcTx["payload"].(string)
		tx.Timestamp = uint64(rpcTx["timestamp"].(float64))
		tx.Fee = int64(rpcTx["fee"].(float64))
		Txs = append(Txs, tx)
	}
	return Txs, nil
}
