package gohfc

import (
	"fmt"
	"github.com/op/go-logging"
	"math/rand"
	"os"
	"time"
)

func getChainCodeObj(args []string) (*ChainCode, error) {
	channelid := handler.client.Channel.ChannelId
	chaincodeName := handler.client.Channel.ChaincodeName
	chaincodeVersion := handler.client.Channel.ChaincodeVersion
	mspId := handler.client.Channel.LocalMspId
	if channelid == "" || chaincodeName == "" || chaincodeVersion == "" || mspId == "" {
		return nil, fmt.Errorf("channelid or ccname or ccver  or mspId is empty")
	}

	chaincode := ChainCode{
		ChannelId: channelid,
		Type:      ChaincodeSpec_GOLANG,
		Name:      chaincodeName,
		Version:   chaincodeVersion,
		Args:      args,
	}

	return &chaincode, nil
}

//设置log级别
func setLogLevel() error {
	modelLevel := handler.client.Log.LogLevel
	modelName := handler.client.Log.LogModelName
	if modelLevel == "" {
		modelLevel = "DEBUG"
	}
	format := logging.MustStringFormatter("%{shortfile} %{time:2006-01-02 15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logLevel, err := logging.LogLevel(modelLevel)
	if err != nil {
		return err
	}
	//map[k]v; eg:  var logger = logging.MustGetLogger("event")
	logging.SetBackend(backendFormatter).SetLevel(logLevel, modelName)
	return nil
}

//解析背书策略
func parsePolicy() error {
	policyOrgs := handler.client.Channel.Orgs
	policyRule := handler.client.Channel.Rule
	if len(policyOrgs) == 0 || policyRule == "" {
		return fmt.Errorf("channl policy config is err")
	}
	for ordname := range handler.client.Orderers {
		orderNames = append(orderNames, ordname)
	}
	for _, v := range handler.client.EventPeers {
		eventName = v.Name
		break
	}
	for _, v := range handler.client.Peers {
		if containsStr(policyOrgs, v.OrgName) {
			orgPeerMap[v.OrgName] = append(orgPeerMap[v.OrgName], v.Name)
			if policyRule == "or" {
				orRulePeerNames = append(orRulePeerNames, v.Name)
			}
		}
	}

	return nil
}

func getSendOrderName() string {
	return orderNames[generateRangeNum(0, len(orderNames))]
}

func getSendPeerName() []string {
	if len(orRulePeerNames) > 0 {
		return []string{orRulePeerNames[generateRangeNum(0, len(orRulePeerNames))]}
	}
	var sendNameList []string
	policyRule := handler.client.Channel.Rule
	if policyRule == "and" {
		for _, peerNames := range orgPeerMap {
			sendNameList = append(sendNameList, peerNames[generateRangeNum(0, len(peerNames))])
			continue
		}
	}

	return sendNameList
}

func generateRangeNum(min, max int) int {
	rand.Seed(time.Now().Unix())
	randNum := rand.Intn(max-min) + min
	return randNum
}

func containsStr(strList []string, str string) bool {
	for _, v := range strList {
		if v == str {
			return true
		}
	}
	return false
}
