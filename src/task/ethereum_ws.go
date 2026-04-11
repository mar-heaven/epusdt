package task

//
//import (
//	"context"
//	"math/big"
//	"strings"
//	"sync/atomic"
//	"time"
//
//	"github.com/assimon/luuu/model/data"
//	"github.com/assimon/luuu/model/mdb"
//	"github.com/assimon/luuu/model/service"
//	"github.com/assimon/luuu/util/log"
//	"github.com/ethereum/go-ethereum"
//	"github.com/ethereum/go-ethereum/common"
//	"github.com/ethereum/go-ethereum/core/types"
//	"github.com/ethereum/go-ethereum/crypto"
//	"github.com/ethereum/go-ethereum/ethclient"
//)
//
//var transferEventSig = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
//
//type ethRecipientSnapshot struct {
//	addrs map[string]struct{}
//}
//
//var ethWatchedRecipients atomic.Pointer[ethRecipientSnapshot]
//
//func storeEthRecipientsFromWallets(wallets []mdb.WalletAddress) int {
//	m := make(map[string]struct{})
//	for _, w := range wallets {
//		a := strings.TrimSpace(w.Address)
//		if !common.IsHexAddress(a) {
//			continue
//		}
//		m[strings.ToLower(common.HexToAddress(a).Hex())] = struct{}{}
//	}
//	ethWatchedRecipients.Store(&ethRecipientSnapshot{addrs: m})
//	return len(m)
//}
//
//func isWatchedEthRecipient(to common.Address) bool {
//	snap := ethWatchedRecipients.Load()
//	if snap == nil || len(snap.addrs) == 0 {
//		return false
//	}
//	_, ok := snap.addrs[strings.ToLower(to.Hex())]
//	return ok
//}
//
////func StartEthereumWebSocketListener(ctx context.Context) {
////	go func() {
////		backoff := time.Second
////		for ctx.Err() == nil {
////			url := config.GetEthereumWsUrl()
////			if url == "" {
////				log.Sugar.Debug("[ETH-WS] ethereum_ws_url empty, disabled")
////				time.Sleep(30 * time.Second)
////				continue
////			}
////			err := runEthereumWebSocket(ctx, url)
////			if err != nil && ctx.Err() == nil {
////				log.Sugar.Warnf("[ETH-WS] subscription ended: %v, reconnect in %s", err, backoff)
////				time.Sleep(backoff)
////				if backoff < 30*time.Second {
////					backoff *= 2
////				}
////				continue
////			}
////			backoff = time.Second
////		}
////	}()
////}
//
//func runEthereumWebSocket(ctx context.Context, url string) error {
//	client, err := ethclient.DialContext(ctx, url)
//	if err != nil {
//		return err
//	}
//	defer client.Close()
//
//	usdt := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
//	usdc := common.HexToAddress("0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
//	refresh := time.NewTicker(2 * time.Minute)
//	defer refresh.Stop()
//
//	for ctx.Err() == nil {
//		innerCtx, cancel := context.WithCancel(ctx)
//
//		wallets, err := data.GetAvailableWalletAddressByNetwork(mdb.NetworkEthereum)
//		if err != nil {
//			cancel()
//			return err
//		}
//		nRec := storeEthRecipientsFromWallets(wallets)
//		if nRec == 0 {
//			cancel()
//			log.Sugar.Debug("[ETH-WS] no valid Ethereum addresses, wait...")
//			select {
//			case <-ctx.Done():
//				return ctx.Err()
//			case <-time.After(30 * time.Second):
//			}
//			continue
//		}
//
//		// used to filter logs in handler, not used in RPC topic condition, because we want to receive all Transfer events and filter in handler to avoid missing logs when recipient list changes.
//		query := ethereum.FilterQuery{
//			Addresses: []common.Address{usdt, usdc},
//		}
//
//		logsCh := make(chan types.Log, 512)
//		sub, err := client.SubscribeFilterLogs(innerCtx, query, logsCh)
//		if err != nil {
//			cancel()
//			return err
//		}
//		log.Sugar.Infof("[ETH-WS] subscribed USDT+USDC Transfer logs; %d watched recipient(s) filtered in handler", nRec)
//
//	subscriptionLoop:
//		for {
//			select {
//			case <-ctx.Done():
//				sub.Unsubscribe()
//				cancel()
//				return ctx.Err()
//			case err := <-sub.Err():
//				sub.Unsubscribe()
//				cancel()
//				return err
//			case <-refresh.C:
//				sub.Unsubscribe()
//				cancel()
//				log.Sugar.Info("[ETH-WS] refreshing subscription (wallet list / reconnect policy)")
//				break subscriptionLoop
//			case vLog := <-logsCh:
//				log.Sugar.Debugf("ewew")
//				handleEthereumTransferLog(ctx, client, vLog)
//			}
//		}
//	}
//	return nil
//}
//
//func handleEthereumTransferLog(ctx context.Context, client *ethclient.Client, vLog types.Log) {
//	defer func() {
//		if r := recover(); r != nil {
//			log.Sugar.Errorf("[ETH-WS] handle log panic: %v", r)
//		}
//	}()
//	if len(vLog.Topics) < 3 {
//		return
//	}
//	if vLog.Topics[0] != transferEventSig {
//		return
//	}
//	if len(vLog.Data) < 32 {
//		return
//	}
//	toAddr := common.BytesToAddress(vLog.Topics[2][12:])
//	//if !isWatchedEthRecipient(toAddr) {
//	//	return
//	//}
//	rawValue := new(big.Int).SetBytes(vLog.Data[:32])
//
//	var blockTsMs int64
//	header, err := client.HeaderByNumber(ctx, big.NewInt(int64(vLog.BlockNumber)))
//	if err != nil {
//		log.Sugar.Warnf("[ETH-WS] HeaderByNumber block=%d: %v, using local time", vLog.BlockNumber, err)
//		blockTsMs = time.Now().UnixMilli()
//	} else {
//		blockTsMs = int64(header.Time) * 1000
//	}
//
//	service.TryProcessEthereumERC20Transfer(vLog.Address, toAddr, rawValue, vLog.TxHash.Hex(), blockTsMs)
//}
