package miner

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"github.com/bazo-blockchain/bazo-miner/p2p"
	"log"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/bazo-blockchain/bazo-miner/protocol"
	"github.com/bazo-blockchain/bazo-miner/storage"
)

var (
	logger              *log.Logger
	blockValidation     = &sync.Mutex{}
	parameterSlice      []Parameters
	activeParameters    *Parameters
	uptodate            bool
	prevBlockIsEpochBlock bool
	slashingDict        = make(map[[64]byte]SlashingProof)
	validatorAccAddress [64]byte
	commPrivKey         *rsa.PrivateKey
	NumberOfShards	    int
	ValidatorShardMap	= make(map[[64]byte]int) // This map keeps track of the validator assignment to the shards; int: shard ID; [64]byte: validator address
	FileConnections		*os.File
)

//Miner entry point
func Init(wallet *ecdsa.PublicKey, commitment *rsa.PrivateKey) error {
	FileConnections,_ = os.OpenFile("hash-prevhash.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	validatorAccAddress = crypto.GetAddressFromPubKey(wallet)
	commPrivKey = commitment

	//Set up logger.
	logger = storage.InitLogger()

	parameterSlice = append(parameterSlice, NewDefaultParameters())
	activeParameters = &parameterSlice[0]

	currentTargetTime = new(timerange)
	target = append(target, 15)

	initialBlock, err := initState() //From here on, every validator should have the same state representation

	FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n",initialBlock.PrevHash[0:15],initialBlock.Hash[0:15]))

	if err != nil {
		return err
	}

	logger.Printf("Active config params:%v", activeParameters)

	/*Sharding Utilities*/
	NumberOfShards = DetNumberOfShards()

	/*For simplicity, the bootstrapping node assigns the validators to the shards*/
	if (p2p.IsBootstrap()){
		/*For simplicity, let the bootstrap server assign all validators and send the map to the other validators*/
		ValidatorShardMap = AssignValidatorsToShards()

		//broadcast the generated map to the other validators
		broadcastValidatorShardMapping(ValidatorShardMap)
	} else {
		//request the mapping from bootstrapping node
		p2p.ValidatorShardMapRequest()
		//Blocking wait
		select {
			case encodedMapping := <-p2p.ValidatorShardMapReq:
				var decodedValMapping map[[64]byte]int
				buffer := bytes.NewBuffer(encodedMapping)
				decoder := gob.NewDecoder(buffer)
				decoder.Decode(&decodedValMapping)

				//Limit waiting time to BLOCKFETCH_TIMEOUT seconds before aborting.
			case <-time.After(BLOCKFETCH_TIMEOUT * time.Second):
				return errors.New("block fetch timeout")
		}
	}


	//Start to listen to network inputs (txs and blocks).
	go incomingData()

	epochMining(initialBlock)

	//mining(initialBlock)

	return nil
}

func InitFirstStart(wallet *ecdsa.PublicKey, commitment *rsa.PrivateKey) error {
	var err error
	FileConnections,err = os.OpenFile("hash-prevhash.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	logger = storage.InitLogger()

	rootAddress := crypto.GetAddressFromPubKey(wallet)

	var rootCommitment [crypto.COMM_KEY_LENGTH]byte
	copy(rootCommitment[:], commitment.N.Bytes())

	genesis := protocol.NewGenesis(rootAddress, rootCommitment)
	storage.WriteGenesis(&genesis)

	logger.Printf("Written Genesis Block: %v\n", genesis.String())

	/*Write First Epoch block chained to the genesis block*/
	initialEpochBlock := protocol.NewEpochBlock([][32]byte{genesis.Hash()},0)
	initialEpochBlock.Hash = initialEpochBlock.HashEpochBlock()
	FirstEpochBlock = initialEpochBlock
	storage.WriteFirstEpochBlock(initialEpochBlock)
	logger.Printf("Written Epoch Block: %v\n", initialEpochBlock.String())

	FileConnections.WriteString(fmt.Sprintf("'GENESIS: %x' -> 'EPOCH BLOCK: %x'\n",[32]byte{},initialEpochBlock.Hash[0:15]))

	return Init(wallet, commitment)
}

func epochMining(initialBlock *protocol.Block) {

	var epochBlock *protocol.EpochBlock
	//currentBlock := newBlock(initialBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, initialBlock.Height + 1)

	lastBlock = initialBlock

	currentBlock := new(protocol.Block)

	currentBlock.Hash = initialBlock.Hash
	currentBlock.Height = initialBlock.Height

	/*constantly watch global blockcount for insertion of next epoch block*/
	for{
		prevBlockIsEpochBlock = false

		if((globalBlockCount + 1) % EPOCH_LENGTH == 0){
			//globalblock count is 1 before the next epoch block. Thus with the last block, create the next epoch block

			//TODO: broadcast hash of lastblock to other validators, this is needed by the epoch block if multiple shards were created
			/*broadcast my last block to the other validators, because epoch block is chained to all last blocks of the shards
			broadcastBlock(lastBlock) */

			//TODO: wait for hashes of the other shards to complete epoch block creation

			epochBlock = protocol.NewEpochBlock([][32]byte{lastBlock.Hash},lastBlock.Height + 1)
			epochBlock.Hash = epochBlock.HashEpochBlock()
			storage.WriteClosedEpochBlock(epochBlock)
			logger.Printf("Inserting EPOCH BLOCK: %v\n",epochBlock.String())

			for _, prevHash := range epochBlock.PrevShardHashes {
				FileConnections.WriteString(fmt.Sprintf("'%x' -> 'EPOCH BLOCK: %x'\n",prevHash[0:15],epochBlock.Hash[0:15]))
			}

			prevBlockIsEpochBlock = true

			mining(epochBlock.Hash,epochBlock.Height)

		} else {
			mining(lastBlock.Hash,lastBlock.Height) // if we are currently not creating epoch block, then continue regular mining
		}
	}
}

//Mining is a constant process, trying to come up with a successful PoW.
func mining(hashPrevBlock [32]byte, heightPrevBlock uint32) {

	currentBlock := newBlock(hashPrevBlock, [crypto.COMM_PROOF_LENGTH]byte{}, heightPrevBlock + 1)

	//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
	//that before start mining a new block we empty the mempool which contains tx data that is likely to be
	//validated with block validation, so we wait in order to not work on tx data that is already validated
	//when we finish the block.
	blockValidation.Lock()
	//nextBlock := newBlock(lastBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, lastBlock.Height+1)
	//currentBlock = nextBlock
	prepareBlock(currentBlock) // In this step, filter tx from mem pool to check if they belong to my shard
	blockValidation.Unlock()

	_, err := storage.ReadAccount(validatorAccAddress)
	if err != nil {
		logger.Printf("%v\n", err)
		time.Sleep(10 * time.Second)
	}

	err = finalizeBlock(currentBlock) //in case another block was mined in the meantime, abort PoS here
	if err != nil {
		logger.Printf("%v\n", err)
	} else {
		logger.Printf("Block mined (%x)\n", currentBlock.Hash[0:8])
	}

	if err == nil {
		broadcastBlock(currentBlock)
		err := validate(currentBlock, false) //here, block is written to closed storage and globalblockcount increased
		if err == nil {
			logger.Printf("Validated block: %vState:\n%v", currentBlock, getState())
			FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n",currentBlock.PrevHash[0:15],currentBlock.Hash[0:15]))
		} else {
			logger.Printf("Received block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err)
		}
	}

	/*for {
		_, err := storage.ReadAccount(validatorAccAddress)
		if err != nil {
			logger.Printf("%v\n", err)
			time.Sleep(10 * time.Second)
			continue
		}

		err = finalizeBlock(currentBlock)
		if err != nil {
			logger.Printf("%v\n", err)
		} else {
			logger.Printf("Block mined (%x)\n", currentBlock.Hash[0:8])
		}

		if err == nil {
			broadcastBlock(currentBlock)
			err := validate(currentBlock, false) //here, block is written to closed storage
			if err == nil {
				logger.Printf("Validated block: %vState:\n%v", currentBlock, getState())
				FileConnections.WriteString(fmt.Sprintf("'%x' -> '%x'\n",currentBlock.PrevHash[0:15],currentBlock.Hash[0:15]))
			} else {
				logger.Printf("Received block (%x) could not be validated: %v\n", currentBlock.Hash[0:8], err)
			}
		}

		//This is the same mutex that is claimed at the beginning of a block validation. The reason we do this is
		//that before start mining a new block we empty the mempool which contains tx data that is likely to be
		//validated with block validation, so we wait in order to not work on tx data that is already validated
		//when we finish the block.
		blockValidation.Lock()
		nextBlock := newBlock(lastBlock.Hash, [crypto.COMM_PROOF_LENGTH]byte{}, lastBlock.Height+1)
		currentBlock = nextBlock
		prepareBlock(currentBlock) // In this step, filter tx from mem pool to check if they belong to my shard
		blockValidation.Unlock()
	}*/
}

func DetNumberOfShards() (numberOfShards int) {
	return int(math.Ceil(float64(GetValidatorsCount()) / float64(VALIDATORS_PER_SHARD)))
}

func AssignValidatorsToShards() map[[64]byte]int {

	shardCount := NumberOfShards

	/*This map denotes which validator is assigned to which shard index*/
	validatorShardAssignment := make(map[[64]byte]int)

	/*Fill validatorAssignedMap with the validators of the current state.
	The bool value indicates whether the validator has been assigned to a shard
	*/
	validatorSlices := make([][64]byte, 0)
	validatorAssignedMap := make(map[[64]byte]bool)
	for _, acc := range storage.State {
		if acc.IsStaking {
			validatorAssignedMap[acc.Address] = false
			validatorSlices = append(validatorSlices, acc.Address)
		}
	}

	/*Iterate over range of shards. At each index, select a random validators
	from the map above and set is bool 'assigned' to TRUE*/
	rand.Seed(time.Now().Unix())
	for j := 1; j <= VALIDATORS_PER_SHARD; j++{
		for i := 1; i <= shardCount; i++ {
			randomValidator := validatorSlices[rand.Intn(len(validatorSlices))]
			if validatorAssignedMap[randomValidator] == false {
				validatorAssignedMap[randomValidator] = true
				validatorShardAssignment[randomValidator] = i
			} else {
				shardCount += 1
			}
		}
	}

	return validatorShardAssignment
}

/*Notes for Epoch check*/
/*if(math.Mod(float64(globalBlockCount),float64(EPOCH_LENGTH + 1)) == 0){

	processingEpochBlock = true
	epochBlock = protocol.NewEpochBlock([][32]byte{currentBlock.Hash},currentBlock.Height + 1)
	epochBlock.Hash = epochBlock.HashEpochBlock()
	storage.WriteClosedEpochBlock(epochBlock)
	logger.Printf("Inserting EPOCH BLOCK: %v\n",epochBlock)

	for _, prevHash := range epochBlock.PrevShardHashes {
	FileConnections.WriteString(fmt.Sprintf("'%x' -> 'EPOCH BLOCK: %x'\n",prevHash[0:15],epochBlock.Hash[0:15]))
	}
}*/