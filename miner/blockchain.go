package miner

import (
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"golang.org/x/crypto/ed25519"
	"log"
	"os"
	"sync"
)

var (
	logger                  *log.Logger
	blockValidation         = &sync.Mutex{}
	payloadMap	            = &sync.Mutex{}
	lastShardMutex			= &sync.Mutex{}
	parameterSlice          []Parameters
	activeParameters        *Parameters
	uptodate                bool
	prevBlockIsEpochBlock   bool
	FirstStartAfterEpoch	bool
	slashingDict            = make(map[[32]byte]SlashingProof)
	validatorAccAddress     [32]byte
	ThisShardID             int // ID of the shard this validator is assigned to
	NumberOfShards          int
	ReceivedBlocksAtHeightX int //This counter is used to sync block heights among shards
	LastShardHashes         [][32]byte // This slice stores the hashes of the last blocks from the other shards, needed to create the next epoch block
	LastShardHashesMap 		= make(map[[32]byte][32]byte)
	ValidatorShardMap       *protocol.ValShardMapping // This map keeps track of the validator assignment to the shards; int: shard ID; [32]byte: validator address
	FileConnections   	       *os.File
	FileConnectionsLog         *os.File
	TransactionPayloadOut 	*protocol.TransactionPayload
	TransactionPayloadReceived 	[]*protocol.TransactionPayload
	TransactionPayloadReceivedMap 	= make(map[[32]byte]*protocol.TransactionPayload)
	TransactionPayloadIn 	[]*protocol.TransactionPayload
	processedTXPayloads		[]int //This slice keeps track of the tx payloads processed from a certain shard
	validatedTXCount		int
	validatedBlockCount		int
	blockStartTime			int64
	blockEndTime			int64
	multisigPubKey      			ed25519.PublicKey
	commPrivKey, rootCommPrivKey	ed25519.PrivateKey
)

//Miner entry point
func Init(validatorWallet, multisigWallet ed25519.PublicKey , rootWallet, validatorCommitment, rootCommitment ed25519.PrivateKey) {
	var err error

	validatorAccAddress = crypto.GetAddressFromPubKeyED(validatorWallet)
	multisigPubKey = multisigWallet
	commPrivKey = validatorCommitment
	rootCommPrivKey = rootCommitment

	//Set up logger.
	logger = storage.InitLogger()

	parameterSlice = append(parameterSlice, NewDefaultParameters())
	activeParameters = &parameterSlice[0]

	//Initialize root key.
	initRootKey(ed25519.PublicKey(rootWallet[32:]))
	if err != nil {
		logger.Printf("Could not create a root account.\n")
	}

	currentTargetTime = new(timerange)
	target = append(target, 15)

	initialBlock, err := initState()
	if err != nil {
		logger.Printf("Could not set up initial state: %v.\n", err)
		return
	}

	logger.Printf("Active config params:%v", activeParameters)

	//Start to listen to network inputs (txs and blocks).
	go incomingData()
	mining(initialBlock)
}

//Mining is a constant process, trying to come up with a successful PoW.
func mining(initialBlock *protocol.Block) {
	currentBlock := newBlock(initialBlock.Hash, [crypto.COMM_PROOF_LENGTH_ED]byte{}, initialBlock.Height+1)

	for {
		err := finalizeBlock(currentBlock)
		if err != nil {
			logger.Printf("%v\n", err)
		} else {
			logger.Printf("Block mined (%x)\n", currentBlock.Hash[0:8])
		}

		if err == nil {
			broadcastBlock(currentBlock)
			err := validate(currentBlock, false)
			if err == nil {
				logger.Printf("Validated block: %vState:\n%v", currentBlock, getState())
			} else {
				logger.Printf("Received block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err)
			}
		}

		//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
		//that before start mining a new block we empty the mempool which contains tx data that is likely to be
		//validated with block validation, so we wait in order to not work on tx data that is already validated
		//when we finish the block.
		blockValidation.Lock()
		nextBlock := newBlock(lastBlock.Hash, [crypto.COMM_PROOF_LENGTH_ED]byte{}, lastBlock.Height+1)
		currentBlock = nextBlock
		prepareBlock(currentBlock)
		blockValidation.Unlock()
	}
}
//At least one root key needs to be set which is allowed to create new accounts.
func initRootKey(rootKey ed25519.PublicKey) error {
	address := crypto.GetAddressFromPubKeyED(rootKey)
	addressHash := protocol.SerializeHashContent(address)

	var commPubKey [crypto.COMM_KEY_LENGTH_ED]byte
	copy(commPubKey[:], rootCommPrivKey[32:])

	rootAcc := protocol.NewAccount(address, [32]byte{}, activeParameters.Staking_minimum, true, commPubKey, nil, nil)
	storage.State[addressHash] = &rootAcc
	storage.RootKeys[addressHash] = &rootAcc

	return nil
}
