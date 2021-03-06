// Copyright 2018 The OEX Team Authors
// This file is part of the OEX project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package sdk

import (
	"math/big"

	"github.com/oexplatform/oexchain/params"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/oexplatform/oexchain/common"
	"github.com/oexplatform/oexchain/rpc"
	"github.com/oexplatform/oexchain/types"
)

// SendRawTransaction send signed tx
func (api *API) SendRawTransaction(rawTx []byte) (common.Hash, error) {
	hash := new(common.Hash)
	err := api.client.Call(hash, "oex_sendRawTransaction", hexutil.Bytes(rawTx))
	return *hash, err
}

// GetCurrentBlock get current block info
func (api *API) GetCurrentBlock(fullTx bool) (map[string]interface{}, error) {
	block := map[string]interface{}{}
	err := api.client.Call(&block, "oex_getCurrentBlock", fullTx)
	return block, err
}

// GetBlockByHash get block info
func (api *API) GetBlockByHash(hash common.Hash, fullTx bool) (map[string]interface{}, error) {
	block := map[string]interface{}{}
	err := api.client.Call(&block, "oex_getBlockByHash", hash, fullTx)
	return block, err
}

// GetBlockByNumber get block info
func (api *API) GetBlockByNumber(number int64, fullTx bool) (map[string]interface{}, error) {
	block := map[string]interface{}{}
	err := api.client.Call(&block, "oex_getBlockByNumber", rpc.BlockNumber(number), fullTx)
	return block, err
}

// GetTransactionByHash get tx info by hash
func (api *API) GetTransactionByHash(hash common.Hash) (*types.RPCTransaction, error) {
	tx := &types.RPCTransaction{}
	err := api.client.Call(tx, "oex_getTransactionByHash", hash)
	return tx, err
}

// GetTransactionReceiptByHash get tx info by hash
func (api *API) GetTransactionReceiptByHash(hash common.Hash) (*types.RPCReceipt, error) {
	receipt := &types.RPCReceipt{}
	err := api.client.Call(receipt, "oex_getTransactionReceipt", hash)
	return receipt, err
}

// GasPrice get gas price
func (api *API) GasPrice() (*big.Int, error) {
	gasprice := big.NewInt(0)
	err := api.client.Call(gasprice, "oex_gasPrice")
	return gasprice, err
}

// GetChainConfig get chain config
func (api *API) GetChainConfig() (*params.ChainConfig, error) {
	cfg := &params.ChainConfig{}
	err := api.client.Call(cfg, "oex_getChainConfig")
	return cfg, err
}

// // GetGenesis get chain config
// func (api *API) GetGenesis() (*params.ChainConfig, error) {
// 	cfg := &params.ChainConfig{}
// 	err := api.client.Call(cfg, "oex_getGenesis")
// 	return cfg, err
// }
