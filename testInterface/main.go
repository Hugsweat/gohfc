package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/op/go-logging"
	"github.com/peersafe/gohfc"
	"github.com/peersafe/gohfc/parseBlock"
	"github.com/peersafe/gohfc/waitTxstatus"
	"strconv"
	"strings"
	"time"
)

type ArrayValue []string

func (s *ArrayValue) String() string {
	return fmt.Sprintf("%v", *s)
}

func (a *ArrayValue) Set(s string) error {
	*a = strings.Split(s, ",")
	return nil
}

var (
	logger = logging.MustGetLogger("testInterface")
)

func main() {
	fmt.Println("################CMD#####################")
	fmt.Println("./testInterface -args=invoke,a,b,1")
	fmt.Println("./testInterface -args=query,a")
	fmt.Println("./testInterface -args=listenfull")
	fmt.Println("./testInterface -args=getblockheight")
	fmt.Println("./testInterface -args=gettxbyid,txid")
	fmt.Println("./testInterface -args=getblockbyno,number")
	fmt.Println("#########################################")
	var args ArrayValue
	flag.Var(&args, "args", "Input array to iterate through.")
	flag.Parse()
	if len(args) == 0 {
		fmt.Println("-----------CMD ERR-------")
		return
	}
	err := gohfc.InitSDK("./client.yaml")
	if err != nil {
		logger.Error(err)
		return
	}
	if err := gohfc.GetHandler().HandleTxStatus(""); err != nil {
		logger.Error(err)
		return
	}

	switch args[0] {
	case "invoke":
		ch := make(chan int)
		for i := 0; i < 100; i++ {
			//go func() {
			res, err := gohfc.GetHandler().SyncInvoke(args, "", "")
			if err != nil {
				logger.Error(err)
				continue
			}

			logger.Debugf("----syncinvoke--TxID--%s\n", res.TxID)
			//}()
		}
		<-ch
	case "query":
		resVal, err := gohfc.GetHandler().Query(args, "", "")
		if err != nil || len(resVal) == 0 {
			logger.Error(err)
			return
		}
		if resVal[0].Error != nil {
			logger.Error(resVal[0].Error)
			return
		}
		if resVal[0].Response.Response.GetStatus() != 200 {
			logger.Error(fmt.Errorf(resVal[0].Response.Response.GetMessage()))
			return
		}
		logger.Debugf("----query--result--%s\n", resVal[0].Response.Response.GetPayload())
	case "getblockheight":
		number, err := gohfc.GetHandler().GetBlockHeight("")
		if err != nil {
			logger.Errorf("getblockheight err = %s", err.Error())
			return
		}
		fmt.Printf("getblockheight----%d\n", number)
	case "gettxbyid":
		tx, err := gohfc.GetHandler().GetTransactionById(args[1], "")
		if err != nil {
			logger.Errorf("gettxbyid err = %s", err.Error())
			return
		}
		str, _ := json.Marshal(tx)
		fmt.Printf("getblockbyno----%s\n", str)
	case "getblockbyno":
		strBlockNum, err := strconv.Atoi(args[1])
		if err != nil {
			logger.Errorf("getblockbyno a err = %s", err.Error())
			return
		}
		block, err := gohfc.GetHandler().GetBlockByNumber(uint64(strBlockNum), "")
		if err != nil {
			logger.Errorf("getblockbyno b err = %s", err.Error())
			return
		}
		plBlock, err := gohfc.GetHandler().ParseCommonBlock(block)
		if err != nil {
			logger.Errorf("getblockbyno c err = %s", err.Error())
			return
		}
		str, _ := json.Marshal(plBlock)
		fmt.Printf("getblockbyno----%s\n", str)
	case "listenfull":
		ch := make(chan parseBlock.Block)
		err := gohfc.GetHandler().EventFullBlock("", -1, ch)
		//err := gohfc.GetHandler().EventFilterBlock("", -1, ch)
		if err != nil {
			logger.Errorf("EventFullBlock err = %s", err.Error())
			return
		}
		aa := func() (uint64, error) {
			return waitTxstatus.GlobalBlockNumber.Get(), nil
		}
		go func() {
			gohfc.GetHandler().JudgeFullEventConnect(aa, gohfc.GetHandler().EventFullBlock, ch)
		}()
		for {
			select {
			case b := <-ch:
				if b.Error != nil {
					logger.Errorf("ListenEventFullBlock err = %s", b.Error.Error())
					continue
				}
				logger.Debugf("------listen block num---%d\n", b.Header.Number)
				if len(b.Transactions) == 0 {
					logger.Debugf("ListenEventFullBlock Config Block BlockNumber= %d, ", b.Header.Number)
				} else {
					waitTxstatus.GlobalBlockNumber.Put(b.Header.Number)
					//aa,_ := json.Marshal(b)
					//logger.Debugf("---%s\n",aa)
				}
			}
		}
	case "listen":
		ch, err := gohfc.GetHandler().Listen("", "")
		if err != nil {
			logger.Error(err)
			return
		}
		for {
			select {
			case b := <-ch:
				logger.Debugf("------listen block num---%d\n", b.Header.Number)
				//aa,_ := json.Marshal(b)
				//logger.Debugf("---%s\n",aa)
			}
		}
	case "checkordconn":
		for {
			ok, err := gohfc.GetHandler().GetOrdererConnect()
			if err != nil {
				logger.Error(err)
				return
			}
			logger.Debugf("the connect is %v", ok)
			time.Sleep(2 * time.Second)
		}
	default:
		flag.PrintDefaults()
	}
	return
}
