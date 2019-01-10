package p2p

import (
	"errors"
)

//Both block and tx requests are handled asymmetricaly, using channels as inter-communication
//All the request in this file are specifically initiated by the miner package
func BlockReq(hash [32]byte) error {

	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(BLOCK_REQ, hash[:])
	sendData(p, packet)
	return nil
}

func ValidatorShardMapRequest() error {
	p := peers.getRandomPeer(PEERTYPE_MINER)

	if p == nil {
		return errors.New("Couldn't get a connection to the bootstrapping node, request not transmitted.")
	}

	packet := BuildPacket(VALIDATOR_SHARD_REQ, nil)
	sendData(p, packet)
	return nil
}

func LastBlockReq() error {

	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(BLOCK_REQ, nil)
	sendData(p, packet)
	return nil
}

func GenesisReq() error {
	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(GENESIS_REQ, nil)
	sendData(p, packet)
	return nil
}

func FirstEpochBlockReq() error {
	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(FIRST_EPOCH_BLOCK_REQ, nil)
	sendData(p, packet)
	return nil
}

func LastEpochBlockReq() error {
	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(LAST_EPOCH_BLOCK_REQ, nil)
	sendData(p, packet)
	return nil
}

func EpochBlockReq(hash [32]byte) error {
	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(EPOCH_BLOCK_REQ, hash[:])
	sendData(p, packet)
	return nil
}

//Request specific transaction
func TxReq(hash [32]byte, reqType uint8) error {

	p := peers.getRandomPeer(PEERTYPE_MINER)
	if p == nil {
		return errors.New("Couldn't get a connection, request not transmitted.")
	}

	packet := BuildPacket(reqType, hash[:])
	sendData(p, packet)
	return nil
}
