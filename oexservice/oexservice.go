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

package oexservice

import (
	"math/big"

	"github.com/ethereum/go-ethereum/log"
	"github.com/oexplatform/oexchain/blockchain"
	"github.com/oexplatform/oexchain/consensus"
	"github.com/oexplatform/oexchain/consensus/dpos"
	"github.com/oexplatform/oexchain/consensus/miner"
	"github.com/oexplatform/oexchain/oexservice/gasprice"
	"github.com/oexplatform/oexchain/node"
	"github.com/oexplatform/oexchain/p2p"
	adaptor "github.com/oexplatform/oexchain/p2p/protoadaptor"
	"github.com/oexplatform/oexchain/params"
	"github.com/oexplatform/oexchain/processor"
	"github.com/oexplatform/oexchain/processor/vm"
	"github.com/oexplatform/oexchain/rpc"
	"github.com/oexplatform/oexchain/rpcapi"
	"github.com/oexplatform/oexchain/txpool"
	"github.com/oexplatform/oexchain/utils/fdb"
)

// FtService implements the oex service.
type FtService struct {
	config       *Config
	chainConfig  *params.ChainConfig
	shutdownChan chan bool // Channel for shutting down the service
	blockchain   *blockchain.BlockChain
	txPool       *txpool.TxPool
	chainDb      fdb.Database // Block chain database
	engine       consensus.IEngine
	miner        *miner.Miner
	p2pServer    *adaptor.ProtoAdaptor
	APIBackend   *APIBackend
}

// New creates a new oexservice object (including the initialisation of the common oexservice object)
func New(ctx *node.ServiceContext, config *Config) (*FtService, error) {
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}

	chainCfg, dposCfg, _, err := blockchain.SetupGenesisBlock(chainDb, config.Genesis)
	if err != nil {
		return nil, err
	}

	ctx.AppendBootNodes(chainCfg.BootNodes)

	ftservice := &FtService{
		config:       config,
		chainDb:      chainDb,
		chainConfig:  chainCfg,
		p2pServer:    ctx.P2P,
		shutdownChan: make(chan bool),
	}

	//blockchain
	vmconfig := vm.Config{
		ContractLogFlag: config.ContractLogFlag,
	}

	ftservice.blockchain, err = blockchain.NewBlockChain(chainDb, config.StatePruning, vmconfig, ftservice.chainConfig, config.BadHashes, config.StartNumber, txpool.SenderCacher)
	if err != nil {
		return nil, err
	}

	// txpool
	if config.TxPool.Journal != "" {
		config.TxPool.Journal = ctx.ResolvePath(config.TxPool.Journal)
	}

	ftservice.txPool = txpool.New(*config.TxPool, ftservice.chainConfig, ftservice.blockchain)

	engine := dpos.New(dposCfg, ftservice.blockchain)
	ftservice.engine = engine

	type bc struct {
		*blockchain.BlockChain
		consensus.IEngine
		*txpool.TxPool
		processor.Processor
	}

	bcc := &bc{
		ftservice.blockchain,
		ftservice.engine,
		ftservice.txPool,
		nil,
	}

	validator := processor.NewBlockValidator(bcc, ftservice.engine)
	txProcessor := processor.NewStateProcessor(bcc, ftservice.engine)

	ftservice.blockchain.SetValidator(validator)
	ftservice.blockchain.SetProcessor(txProcessor)

	bcc.Processor = txProcessor
	ftservice.miner = miner.NewMiner(bcc)
	ftservice.miner.SetDelayDuration(config.Miner.Delay)
	ftservice.miner.SetCoinbase(config.Miner.Name, config.Miner.PrivateKeys)
	ftservice.miner.SetExtra([]byte(config.Miner.ExtraData))
	if config.Miner.Start {
		ftservice.miner.Start(false)
	}

	ftservice.APIBackend = &APIBackend{ftservice: ftservice}

	ftservice.SetGasPrice(ftservice.TxPool().GasPrice())
	return ftservice, nil
}

// APIs return the collection of RPC services the oexservice package offers.
func (fs *FtService) APIs() []rpc.API {
	return rpcapi.GetAPIs(fs.APIBackend)
}

// Start implements node.Service, starting all internal goroutines.
func (fs *FtService) Start() error {
	log.Info("start oex service...")
	return nil
}

// Stop implements node.Service, terminating all internal goroutine
func (fs *FtService) Stop() error {
	fs.miner.Stop()
	fs.blockchain.Stop()
	fs.txPool.Stop()
	fs.chainDb.Close()
	close(fs.shutdownChan)
	log.Info("oexservice stopped")
	return nil
}

func (fs *FtService) GasPrice() *big.Int {
	return fs.txPool.GasPrice()
}

func (fs *FtService) SetGasPrice(gasPrice *big.Int) bool {
	fs.config.GasPrice.Default = new(big.Int).SetBytes(gasPrice.Bytes())
	fs.APIBackend.gpo = gasprice.NewOracle(fs.APIBackend, fs.config.GasPrice)
	fs.txPool.SetGasPrice(new(big.Int).SetBytes(gasPrice.Bytes()))
	return true
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (fdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (s *FtService) BlockChain() *blockchain.BlockChain { return s.blockchain }
func (s *FtService) TxPool() *txpool.TxPool             { return s.txPool }
func (s *FtService) Engine() consensus.IEngine          { return s.engine }
func (s *FtService) ChainDb() fdb.Database              { return s.chainDb }
func (s *FtService) Protocols() []p2p.Protocol          { return nil }
