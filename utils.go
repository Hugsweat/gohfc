package gohfc

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/common/channelconfig"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/op/go-logging"
)

func getChainCodeObj(args []string, channelName, chaincodeName string) (*ChainCode, error) {
	if len(channelName) == 0 {
		channelName = handler.client.Channel.ChannelId
	}
	if len(chaincodeName) == 0 {
		chaincodeName = handler.client.Channel.ChaincodeName
	}
	mspId := handler.client.Channel.LocalMspId
	if channelName == "" || chaincodeName == "" || mspId == "" {
		return nil, fmt.Errorf("channelName or chaincodeName or mspId is empty")
	}

	chaincode := ChainCode{
		ChannelId: channelName,
		Type:      ChaincodeSpec_GOLANG,
		Name:      chaincodeName,
		Args:      args,
	}

	return &chaincode, nil
}

//设置log级别
func SetLogLevel(level, name string) error {
	format := logging.MustStringFormatter("%{shortfile} %{time:2006-01-02 15:04:05.000} [%{module}] %{level:.4s} : %{message}")
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, format)
	logLevel, err := logging.LogLevel(level)
	if err != nil {
		return err
	}
	logging.SetBackend(backendFormatter).SetLevel(logLevel, name)
	logger.Debugf("SetLogLevel level: %s, levelName: %s\n", level, name)
	return nil
}

func GetLogLevel(name string) string {
	level := logging.GetLevel(name).String()
	logger.Debugf("GetLogLevel level: %s, LogModule: %s\n", level, name)
	return level
}

/*
//解析背书策略
func parsePolicy() error {
	policyOrgs := handler.client.Channel.Orgs
	policyRule := handler.client.Channel.Rule
	if len(policyOrgs) == 0 || policyRule == "" {
		for _, v := range handler.client.Peers {
			peerNames = append(peerNames, v.Name)
		}
	}
	for ordname := range handler.client.Orderers {
		orderNames = append(orderNames, ordname)
	}
	for _, v := range handler.client.EventPeers {
		eventName = v.Name
		break
	}
	if len(policyOrgs) > 0 {
		for _, v := range handler.client.Peers {
			if containsStr(policyOrgs, v.OrgName) {
				orgPeerMap[v.OrgName] = append(orgPeerMap[v.OrgName], v.Name)
				if policyRule == "or" {
					orRulePeerNames = append(orRulePeerNames, v.Name)
				}
			}
		}
	}

	return nil
}
*/
func getSendOrderName() string {
	return orderNames[generateRangeNum(0, len(orderNames))]
}

func getSendPeerName() []string {
	if len(orRulePeerNames) > 0 {
		return []string{orRulePeerNames[generateRangeNum(0, len(orRulePeerNames))]}
	}
	if len(peerNames) > 0 {
		return peerNames
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

func GetAnchorPeersFromBlock(block *common.Block) (map[string][]*peer.AnchorPeer, error) {
	orgAnchorPeers := make(map[string][]*peer.AnchorPeer)

	if block == nil {
		return orgAnchorPeers, errors.New("block is nil")
	}

	if !utils.IsConfigBlock(block) {
		return orgAnchorPeers, errors.New("the block is not a config block")
	}

	envelope, err := utils.ExtractEnvelope(block, 0)
	if err != nil {
		return orgAnchorPeers, errors.New("can not extract the envelope info")
	}

	configEnv := &common.ConfigEnvelope{}
	_, err = utils.UnmarshalEnvelopeOfType(envelope, common.HeaderType_CONFIG, configEnv)
	if err != nil {
		return orgAnchorPeers, errors.New("can not unmarshal envelope to config envelope")
	}

	appGroup := configEnv.Config.ChannelGroup.Groups["Application"]
	for key, conGroupValue := range appGroup.Groups {
		anchorValue := conGroupValue.Values[channelconfig.AnchorPeersKey]
		anchorPeers := &peer.AnchorPeers{}
		err = proto.Unmarshal(anchorValue.Value, anchorPeers)
		if err != nil {
			logger.Errorf("unmarshal anchor peers for %s failed %s", key, err.Error())
			continue
		}
		orgAnchorPeers[key] = anchorPeers.AnchorPeers
	}

	return orgAnchorPeers, nil
}
