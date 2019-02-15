package protocol

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
)

type StateTransition struct {
	RelativeStateChange			map[[32]byte]*RelativeAccount
	Height						int
	ShardID						int
}

type RelativeAccount struct {
	Address            [32]byte              // 64 Byte
	Issuer             [32]byte              // 64 Byte
	Balance            int64                // 8 Byte
	TxCnt              int32                // 4 Byte
	IsStaking          bool                  // 1 Byte
	CommitmentKey      [crypto.COMM_KEY_LENGTH_ED]byte// represents the modulus N of the RSA public key
	StakingBlockHeight int32                // 4 Byte
	Contract           []byte                // Arbitrary length
	ContractVariables  []ByteArray           // Arbitrary length
}

func NewStateTransition(stateChange map[[32]byte]*RelativeAccount, height int, shardid int) *StateTransition {
	newTransition := StateTransition{
		stateChange,
		height,
		shardid,
	}

	return &newTransition
}

func NewRelativeAccount(address [32]byte,
	issuer [32]byte,
	balance int64,
	isStaking bool,
	commitmentKey [crypto.COMM_KEY_LENGTH_ED]byte,
	contract []byte,
	contractVariables []ByteArray) RelativeAccount {

	newAcc := RelativeAccount{
		address,
		issuer,
		balance,
		0,
		isStaking,
		commitmentKey,
		0,
		contract,
		contractVariables,
	}

	return newAcc
}

func (st *StateTransition) HashTransition() [32]byte {
	if st == nil {
		return [32]byte{}
	}

	stHash := struct {
		RelativeStateChange				  map[[32]byte]*RelativeAccount
		Height				  			  int
		ShardID							  int
	}{
		st.RelativeStateChange,
		st.Height,
		st.ShardID,
	}
	return SerializeHashContent(stHash)
}


func (acc *RelativeAccount) Hash() [32]byte {
	if acc == nil {
		return [32]byte{}
	}

	return SerializeHashContent(acc.Address)
}

func (st *StateTransition) EncodeTransition() []byte {
	if st == nil {
		return nil
	}

	encoded := StateTransition{
		RelativeStateChange:		st.RelativeStateChange,
		Height:						st.Height,
		ShardID:					st.ShardID,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (*StateTransition) DecodeTransition(encoded []byte) (st *StateTransition) {
	var decoded StateTransition
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (acc *RelativeAccount) Encode() []byte {
	if acc == nil {
		return nil
	}

	encoded := RelativeAccount{
		Address:            acc.Address,
		Issuer:             acc.Issuer,
		Balance:            acc.Balance,
		TxCnt:              acc.TxCnt,
		IsStaking:          acc.IsStaking,
		CommitmentKey:   	acc.CommitmentKey,
		StakingBlockHeight: acc.StakingBlockHeight,
		Contract:           acc.Contract,
		ContractVariables:  acc.ContractVariables,
	}

	buffer := new(bytes.Buffer)
	gob.NewEncoder(buffer).Encode(encoded)
	return buffer.Bytes()
}

func (*RelativeAccount) Decode(encoded []byte) (acc *RelativeAccount) {
	var decoded RelativeAccount
	buffer := bytes.NewBuffer(encoded)
	decoder := gob.NewDecoder(buffer)
	decoder.Decode(&decoded)
	return &decoded
}

func (acc RelativeAccount) String() string {
	addressHash := acc.Hash()
	return fmt.Sprintf(
		"Hash: %x, " +
			"Address: %x, " +
			"Issuer: %x, " +
			"TxCnt: %v, " +
			"Balance: %v, " +
			"IsStaking: %v, " +
			"CommitmentKey: %x, " +
			"StakingBlockHeight: %v, " +
			"Contract: %v, " +
			"ContractVariables: %v",
		addressHash[0:8],
		acc.Address[0:8],
		acc.Issuer[0:8],
		acc.TxCnt,
		acc.Balance,
		acc.IsStaking,
		acc.CommitmentKey[0:8],
		acc.StakingBlockHeight,
		acc.Contract,
		acc.ContractVariables)
}

