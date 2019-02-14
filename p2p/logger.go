package p2p

import (
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/storage"
	"log"
	"os"
	"strings"
)

var (
	LogMapping map[uint8]string
	logger     *log.Logger
	FileConnectionsLog         *os.File
)

func InitLogging() {
	logger = storage.InitLogger()
	FileConnectionsLog, _ = os.OpenFile(fmt.Sprintf("hlog-for-%v.txt",strings.Split(Ipport, ":")[0]), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	//Instead of logging just the integer, we log the corresponding semantic meaning, makes scrolling through
	//the log file more comfortable
	LogMapping = make(map[uint8]string)
	LogMapping[1] = "FUNDSTX_BRDCST"
	LogMapping[2] = "ACCTX_BRDCST"
	LogMapping[3] = "CONFIGTX_BRDCST"
	LogMapping[4] = "STAKETX_BRDCST"
	LogMapping[5] = "VERIFIEDTX_BRDCST"
	LogMapping[6] = "BLOCK_BRDCST"
	LogMapping[7] = "BLOCK_HEADER_BRDCST"
	LogMapping[8] = "TX_BRDCST_ACK"

	LogMapping[10] = "FUNDSTX_REQ"
	LogMapping[11] = "CONTRACTTX_REQ"
	LogMapping[12] = "CONFIGTX_REQ"
	LogMapping[13] = "STAKETX_REQ"
	LogMapping[14] = "BLOCK_REQ"
	LogMapping[15] = "BLOCK_HEADER_REQ"
	LogMapping[16] = "ACC_REQ"
	LogMapping[17] = "ROOTACC_REQ"
	LogMapping[18] = "INTERMEDIATE_NODES_REQ"
	LogMapping[19] = "GENESIS_REQ"

	LogMapping[20] = "FUNDSTX_RES"
	LogMapping[21] = "CONTRACTTX_RES"
	LogMapping[22] = "CONFIGTX_RES"
	LogMapping[23] = "STAKETX_RES"
	LogMapping[24] = "BlOCK_RES"
	LogMapping[25] = "BlOCK_HEADER_RES"
	LogMapping[26] = "ACC_RES"
	LogMapping[27] = "ROOTACC_RES"
	LogMapping[28] = "INTERMEDIATE_NODES_RES"
	LogMapping[29] = "GENESIS_RES"

	LogMapping[30] = "NEIGHBOR_REQ"

	LogMapping[40] = "NEIGHBOR_RES"

	LogMapping[50] = "TIME_BRDCST"

	LogMapping[100] = "MINER_PING"
	LogMapping[101] = "MINER_PONG"
	LogMapping[102] = "CLIENT_PING"
	LogMapping[103] = "CLIENT_PONG"

	LogMapping[110] = "NOT_FOUND"

	LogMapping[120] = "STATE_REQ"
	LogMapping[121] = "STATE_RES"

	LogMapping[122] = "FIRST_EPOCH_BLOCK_REQ"
	LogMapping[123] = "FIRST_EPOCH_BLOCK_RES"
	LogMapping[124] = "EPOCH_BLOCK_REQ"
	LogMapping[125] = "EPOCH_BLOCK_RES"
	LogMapping[126] = "VALIDATOR_SHARD_BRDCST"
	LogMapping[127] = "VALIDATOR_SHARD_REQ"
	LogMapping[128] = "VALIDATOR_SHARD_RES"
	LogMapping[129] = "EPOCH_BLOCK_BRDCST"
	LogMapping[130] = "LAST_EPOCH_BLOCK_REQ"
	LogMapping[131] = "LAST_EPOCH_BLOCK_RES"
	LogMapping[132] = "TX_PAYLOAD_BRDCST"
	LogMapping[133] = "STATE_TRANSITION_BRDCST"
}
