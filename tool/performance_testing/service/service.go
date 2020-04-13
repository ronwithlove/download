package service

import (
	"context"
	"fmt"
	"github.com/dappley/go-dappley/common"
	"github.com/dappley/go-dappley/core/account"
	transactionpb "github.com/dappley/go-dappley/core/transaction/pb"
	"github.com/dappley/go-dappley/logic"
	rpcpb "github.com/dappley/go-dappley/rpc/pb"
	util_ron "github.com/dappley/go-dappley/tool/performance_testing/util"
	logger "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
	"time"
)

const (
	IP              = "127.0.0.1"
	port            = "50051"//node1的端口
	amountFromminer = 100000 //矿工一次可挖1000万
	amount          = 1
	minerPrivKey    = "300c0338c4b0d49edc66113e3584e04c6b907f9ded711d396d522aae6a79be1a" //node1的私钥
	tps				= 1
)

type Service struct {
	conn rpcpb.RpcServiceClient
	connAd rpcpb.AdminServiceClient
}

func NewServiceClient() *Service {
	conn, err := grpc.Dial(fmt.Sprint(IP, ":", port), grpc.WithInsecure())
	if err != nil {
		fmt.Println("网络异常", err)
		return nil
	}
	//defer conn.Close()
	return &Service{
		conn:rpcpb.NewRpcServiceClient(conn),
		connAd:rpcpb.NewAdminServiceClient(conn),
	}
}

//开始Go程交易
func  (ser *Service) StartTransactionGoroutine(g int){
	go func() {
		minerAccount := account.NewAccountByPrivateKey(minerPrivKey)
		fromAccount, _ := createAccount()
		ticker := time.NewTicker(time.Second*2)
		defer ticker.Stop()
	OuterLoop:
		for {
			select {
			case t := <-ticker.C:
				fmt.Println("Waiting miner to get enough token...")
				if ser.getBalance(minerAccount) > amountFromminer {
					fmt.Println(t.Format("2006-01-02 15:04:05"))
					ser.minerSendTokenToAccount(amountFromminer, *fromAccount) //这里也会有个可能，当查询时候矿工钱有很多，但是等到打款的时候，打给太多人了，钱不够了
					break OuterLoop
				}
			}
		}

		toAccount, _ := createAccount()
		ticker2 := time.NewTicker(time.Second * 1/tps) //定时1秒
		defer ticker2.Stop()

		for {
			select {
			case t := <-ticker2.C:
				fmt.Println("Goroutine: ",g)
				fmt.Println(t.Format("2006-01-02 15:04:05"))
				fmt.Println("From Address:", util_ron.GetAccountName(fromAccount)," balance:", ser.getBalance(fromAccount))
				fmt.Println("To Address:", util_ron.GetAccountName(toAccount)," balance:", ser.getBalance(toAccount))
				//当我问服务器要余额时候，余额还有，但是这个时候
				//之前的交易还没被加到链上，那么当我再次交易时候，很可能服务器余额就不满足了，
				//这个时候就报错了，这里要处理下，此笔交易会被忽略。The transaction was abandoned
				//可以通过本地也保存一个balance来避免这个问题
				//还有可能是上面已经执行来矿工打款过来，但是实际还没到账，就执行下面else的来逻辑
				//这里会报错：Error: transaction verification failed 可以忽略，只是交易不成功，接下来有钱了会继续
				if ser.getBalance(fromAccount) > amount {
					ser.sendToken(amount, fromAccount, toAccount)
				} else {
					ser.sendToken(ser.getBalance(toAccount), toAccount, fromAccount)
				}
				fmt.Println("")
			}
		}
	}()
}



//本地创建账户
func createAccount() (*account.Account, error) {
	account, err := logic.CreateAccountWithPassphrase("123")
	if err != nil {
		logger.WithError(err).Error("Cannot create new account.")
	}
	logger.WithFields(logger.Fields{
		"address": account.GetAddress(),
	}).Info("Account is created")
	return account, err
}

//从矿工拿钱，得等到挖出一个快，矿工才有钱，todo:要等待，得等block到第二块的时候
func (rpc *Service)minerSendTokenToAccount(amount uint64, account account.Account) {
	sendFromMinerRequest := &rpcpb.SendFromMinerRequest{To: account.GetAddress().String(), Amount: common.NewAmount(amount).Bytes()}

	//通过句柄调用函数：rpc RpcSendFromMiner (SendFromMinerRequest) returns (SendFromMinerResponse) {}，
	_, err := rpc.connAd.RpcSendFromMiner(context.Background(), sendFromMinerRequest) //SendFromMinerResponse里啥都没返，就不接收了
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", err.Error())
		}
		return
	}
	fmt.Println(amount, " has been sent to FromAddress.")
}

//付款
func (ser *Service)sendToken(amount uint64, fromAccount, toAccount *account.Account) {
	//从服务器得到响应，包含指定账户地址的utxo信息
	response, err := ser.conn.RpcGetUTXO(context.Background(), &rpcpb.GetUTXORequest{
		Address: fromAccount.GetAddress().String()})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}

	//创建交易
	tx, err := util_ron.CreateTransaction(response, amount, fromAccount, toAccount)
	if err != nil {
		fmt.Println("The transaction was abandoned.")
		return
	}
	//发送交易请求
	sendTransactionRequest := &rpcpb.SendTransactionRequest{Transaction: tx.ToProto().(*transactionpb.Transaction)}
	_, err = ser.conn.RpcSendTransaction(context.Background(), sendTransactionRequest)
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		return
	}

	fmt.Println("New transaction is sent! ")
}

//得到指定账户的余额
func (ser *Service)getBalance(account *account.Account) uint64 {
	response, err := ser.conn.RpcGetBalance(context.Background(), &rpcpb.GetBalanceRequest{Address: account.GetAddress().String()})
	if err != nil {
		switch status.Code(err) {
		case codes.Unavailable:
			fmt.Println("Error: server is not reachable!")
		default:
			fmt.Println("Error:", status.Convert(err).Message())
		}
		os.Exit(1)
	}
	return uint64(response.GetAmount())
}


