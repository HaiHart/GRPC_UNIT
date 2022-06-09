package eth

import (
	"fmt"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/eth/protocols/eth"
)

func handleGetBlockHeaders(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleGetBlockHeaders66(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func answerGetBlockHeaders(backend Backend, query *eth.GetBlockHeadersPacket, peer *Peer) ([]*ethtypes.Header, error) {
	return nil, nil
}

func handleGetBlockBodies(backend Backend, msg Decoder, peer *Peer) error {
	return nil
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

func handlePooledTransactions(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handlePooledTransactions66(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleNewPooledTransactionHashes(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleNewBlockHashes(backend Backend, msg Decoder, peer *Peer) error {
	var blockHashes eth.NewBlockHashesPacket
	if err := msg.Decode(blockHashes); err != nil {
		return fmt.Errorf("could not decode message: %v: %v", msg, err)
	}

	updatePeerHeadFromNewHashes(blockHashes, peer)
	return backend.Handle(peer, &blockHashes)
}

func handleBlockHeaders(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleBlockHeaders66(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleBlockBodies(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func handleBlockBodies66(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}

func updatePeerHeadFromNewHashes(newBlocks eth.NewBlockHashesPacket, peer *Peer) {
	if len(newBlocks) > 0 {
		maxHeight := newBlocks[0].Number
		hash := newBlocks[0].Hash
		for _, block := range newBlocks[1:] {
			number := block.Number
			if number > maxHeight {
				maxHeight = number
				hash = block.Hash
			}
		}
		peer.UpdateHead(maxHeight, hash)
	}
}

func handleUnimplemented(backend Backend, msg Decoder, peer *Peer) error {
	return nil
}
