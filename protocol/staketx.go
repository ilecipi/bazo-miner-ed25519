package protocol

import (
	"encoding/binary"
	"fmt"
	"github.com/bazo-blockchain/bazo-miner/crypto"
	"golang.org/x/crypto/ed25519"
)

const (
	STAKETX_SIZE = 138 + crypto.COMM_KEY_LENGTH_ED
)

//when we broadcast transactions we need a way to distinguish with a type

type StakeTx struct {
	Header        byte                  // 1 Byte
	Fee           uint64                // 8 Byte
	IsStaking     bool                  // 1 Byte
	Account       [32]byte              // 64 Byte
	Sig           [64]byte              // 64 Byte
	CommitmentKey [crypto.COMM_KEY_LENGTH_ED]byte // the modulus N of the RSA public key
}

func ConstrStakeTx(header byte, fee uint64, isStaking bool, account [32]byte, signKey ed25519.PrivateKey, commPubKey ed25519.PublicKey) (tx *StakeTx, err error) {

	tx = new(StakeTx)

	tx.Header = header
	tx.Fee = fee
	tx.IsStaking = isStaking
	tx.Account = account

	tx.CommitmentKey = crypto.GetAddressFromPubKeyED(commPubKey)

	txHash := tx.Hash()

	sign := ed25519.Sign(signKey, txHash[:])

	copy(tx.Sig[:],sign[:])

	if err != nil {
		return nil, err
	}


	return tx, nil
}

func (tx *StakeTx) Hash() (hash [32]byte) {
	if tx == nil {
		//is returning nil better?
		return [32]byte{}
	}

	txHash := struct {
		Header     byte
		Fee        uint64
		IsStaking  bool
		Account    [32]byte
		CommKey    [32]byte
	}{
		tx.Header,
		tx.Fee,
		tx.IsStaking,
		tx.Account,
		tx.CommitmentKey,
	}

	return SerializeHashContent(txHash)
}

//when we serialize the struct with binary.Write, unexported field get serialized as well, undesired
//behavior. Therefore, writing own encoder/decoder
func (tx *StakeTx) Encode() (encodedTx []byte) {
	if tx == nil {
		return nil
	}

	var fee [8]byte
	var isStaking byte

	binary.BigEndian.PutUint64(fee[:], tx.Fee)

	if tx.IsStaking == true {
		isStaking = 1
	} else {
		isStaking = 0
	}

	encodedTx = make([]byte, STAKETX_SIZE)

	encodedTx[0] = tx.Header
	copy(encodedTx[1:9], fee[:])
	encodedTx[9] = isStaking
	copy(encodedTx[10:74], tx.Account[:])
	copy(encodedTx[74:138], tx.Sig[:])
	copy(encodedTx[138:138+crypto.COMM_KEY_LENGTH_ED], tx.CommitmentKey[:])

	return encodedTx
}

func (*StakeTx) Decode(encodedTx []byte) (tx *StakeTx) {
	tx = new(StakeTx)

	if len(encodedTx) != STAKETX_SIZE {
		return nil
	}

	var isStakingAsByte byte

	tx.Header = encodedTx[0]
	tx.Fee = binary.BigEndian.Uint64(encodedTx[1:9])
	isStakingAsByte = encodedTx[9]
	copy(tx.Account[:], encodedTx[10:74])
	copy(tx.Sig[:], encodedTx[74:138])
	copy(tx.CommitmentKey[:], encodedTx[138:138+crypto.COMM_KEY_LENGTH_ED])

	if isStakingAsByte == 0 {
		tx.IsStaking = false
	} else {
		tx.IsStaking = true
	}

	return tx
}

func (tx *StakeTx) TxFee() uint64 { return tx.Fee }
func (tx *StakeTx) Size() uint64  { return STAKETX_SIZE }

func (tx StakeTx) String() string {
	return fmt.Sprintf(
		"\nHeader: %x\n"+
			"Fee: %v\n"+
			"IsStaking: %v\n"+
			"Account: %x\n"+
			"Sig: %x\n"+
			"CommitmentKey: %x\n",
		tx.Header,
		tx.Fee,
		tx.IsStaking,
		tx.Account[0:8],
		tx.Sig[0:8],
		tx.CommitmentKey[0:8],
	)
}
