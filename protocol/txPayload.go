package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type TransactionPayload struct {
	ContractTxData  [][32]byte
	FundsTxData  	[][32]byte
	ConfigTxData 	[][32]byte
	StakeTxData  	[][32]byte
}

func NewTransactionPayload(prevHash [32]byte, height uint32) *TransactionPayload {
	newPayload := TransactionPayload{
		ContractTxData: [][32]byte{},
		FundsTxData: [][32]byte{},
		ConfigTxData: [][32]byte{},
		StakeTxData: [][32]byte{},
	}

	return &newPayload
}

func (txPayload *TransactionPayload) HashPayload() [32]byte {
	if txPayload == nil {
		return [32]byte{}
	}

	payloadHash := struct {
		contractTxData           [][32]byte
		fundsTxData              [][32]byte
		configTxData             [][32]byte
		stakeTxData              [][32]byte
	}{
		txPayload.ContractTxData,
		txPayload.FundsTxData,
		txPayload.ConfigTxData,
		txPayload.StakeTxData,
	}
	return SerializeHashContent(payloadHash)
}

func (txPayload *TransactionPayload) GetPayloadSize() int {
	size :=
		len(txPayload.ContractTxData) + len(txPayload.FundsTxData) + len(txPayload.ConfigTxData) + len(txPayload.StakeTxData)
	return size
}

func (txPayload *TransactionPayload) EncodePayload() []byte {
	if txPayload == nil {
		return nil
	}

	encoded := TransactionPayload{
		ContractTxData:        txPayload.ContractTxData,
		FundsTxData:           txPayload.FundsTxData,
		ConfigTxData:          txPayload.ConfigTxData,
		StakeTxData:           txPayload.StakeTxData,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (txPayload *TransactionPayload) DecodePayload(encoded []byte) (txP *TransactionPayload) {
	if encoded == nil {
		return nil
	}

	var decoded TransactionPayload
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (txPayload TransactionPayload) StringPayload() string {
	return fmt.Sprintf("\nHash: %x\n"+
		"TX Payload Hashes: %v\n",
		txPayload.HashPayload()[0:8],
		txPayload.PayloadToString(),
	)
}

func (txPayload TransactionPayload) PayloadToString() (payload string) {

	payload += "=== Contract Tx ==="

	for _, tx := range txPayload.ContractTxData {
		payload += fmt.Sprintf("\n%x", tx)
	}

	payload += "\n=== Funds Tx ==="

	for _, tx := range txPayload.FundsTxData {
		payload += fmt.Sprintf("\n%x", tx)
	}

	payload += "\n=== Config Tx ==="

	for _, tx := range txPayload.ConfigTxData {
		payload += fmt.Sprintf("\n%x", tx)
	}

	payload += "\n=== Stake Tx ==="

	for _, tx := range txPayload.StakeTxData {
		payload += fmt.Sprintf("\n%x", tx)
	}


	return payload
}