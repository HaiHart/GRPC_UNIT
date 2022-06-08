package eth

import (
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
)

func handleGetBlockHeaders66(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func answerGetBlockHeaders(backend Backend, query *eth.GetBlockHeadersPacket, peer *Peer) ([]*ethtypes.Header, error) {
	return nil, nil
}

func handleGetBlockBodies66(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func answerGetBlockBodies(backend Backend, query eth.GetBlockBodiesPacket) ([]*eth.BlockBody, error) {
	return nil, nil
}

func handleNewBlockMsg(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleTransactions(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handlePooledTransactions66(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleNewPooledTransactionHashes(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleNewBlockHashes(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleBlockHeaders66(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleBlockBodies66(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleUnimplemented(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}
