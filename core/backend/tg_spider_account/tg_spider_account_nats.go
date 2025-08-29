package tg_spider_account

import (
	"errors"
	"fmt"
	"github.com/bytedance/sonic"
	nats2 "github.com/nats-io/nats.go"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"jarvis/middleware/mq/nats"
	"operate-backend/core/structure"
	"strconv"
	"strings"
	"sync"
)

const (
	TGSpiderAccountDistributionSubject = "TGSpider.Account.Distribution" // 账号分发主题
	TGSpiderOBESubjectPrefix           = "TGSpider.OBE."                 // 后台主题
	TGSpiderSSSubjectPrefix            = "TGSpider.SS."                  // 爬虫主题
)

const (
	OBEActAck       = "Ack"
	OBEActCode      = "Code"
	OBEActSession   = "Session"
	OBEActReady     = "Ready"
	OBEActOrderDone = "OrderDone"
	OBEActStoped    = "Stoped"
	SSActCode       = "Code"
	SSActOrder      = "Order"
	SSActStop       = "Stop"
)

type StartInfo struct {
	ID      uint   `json:"id"`       // 通讯唯一ID
	Phone   string `json:"phone"`    // 手机号码
	Code    string `json:"code"`     // 验证码
	TFAPwd  string `json:"tfa_pwd"`  // 2FA密码
	AppID   uint64 `json:"app_id"`   // AppID
	AppHash string `json:"app_hash"` // AppHash
	Session string `json:"session"`  // 缓存
}

var (
	_subscription *nats2.Subscription
	_idMap        = new(sync.Map)
	_startChannel = make(chan StartInfo, 1000)
	_stopChannel  = make(chan uint, 1000)
	_close        = make(chan struct{})
	_done         = make(chan struct{})
)

func InitSubscribe() error {
	if nats.Instance() == nil {
		return errors.New("nats instance is nil")
	}

	subscription, err := nats.Instance().Subscribe(TGSpiderOBESubjectPrefix+">", Handler)
	if err != nil {
		return err
	}

	_subscription = subscription

	// 下发 -> Account.Distribution : 1
	// 										->  SS_01 抢到后就往 TGSpider.OBE.1 发送一条收到
	// OBE 15秒内收到 ACK 就往 TGSpider.SS.1 发送消息

	// TGSpider.OBE.1.Ack			// 爬虫确认收到该账号信息
	// TGSpider.OBE.1.Code			// 爬虫索要验证码
	// TGSpider.OBE.1.Session		// 爬虫发送Session缓存过来
	// TGSpider.OBE.1.Ready			// 爬虫告知已经准备好接收任务
	// TGSpider.OBE.1.OrderDone 	// 爬虫确认完成当前任务
	// TGSpider.OBE.1.Stoped		// 爬虫告知已经停止
	// TGSpider.SS.1.Code			// 后台下发验证码
	// TGSpider.SS.1.Order			// 后台下发爬虫任务
	// TGSpider.SS.1.Stop			// 后台下发停止指令(该指令会在爬虫完成当前任务后才会生效)

	return nil
}

func NatsLoop() {
	defer close(_done)

	logger.App().Infof("================================= TGSpiderNats start =================================")
	defer logger.App().Infof("================================= TGSpiderNats stop =================================")

	for {
		done := false

		select {
		case <-_close:
			{
				done = true
			}
		case info, ok := <-_startChannel:
			{
				if !ok {
					done = true
					break
				}

				if info.ID == 0 {
					logger.App().Errorf("id of start info is zero: %+v", info)
					continue
				}

				data, err := sonic.Marshal(&info)
				if err != nil {
					logger.App().Errorf("marshal start info err: %+v-%s", info, err.Error())
					continue
				}

				if err = nats.Instance().Publish(TGSpiderAccountDistributionSubject, data[:]); err != nil {
					logger.App().Errorf("publish start info err: %s-%s", string(data), err.Error())
					continue
				}

				_idMap.Store(fmt.Sprintf("%d", info.ID), 0) // 存入，0-未得到响应

				if err = structure.UpdateProcess(info.ID, 2); err != nil {
					logger.App().Errorf("update process error: %d-%d-%s", info.ID, 2, err.Error())
					continue
				}

				logger.App().Infof("publish start tg spider account info success : %s", string(data))
			}
		case id, ok := <-_stopChannel:
			{
				if !ok {
					done = true
					break
				}

				value, exist := _idMap.Load(fmt.Sprintf("%d", id))
				if !exist {
					logger.App().Errorf("id not register : %d", id)
					continue
				}
				status := value.(int)

				if status == 0 {
					logger.App().Errorf("id not ack : %d", id)
					continue
				}

				if err := nats.Instance().Publish(
					TGSpiderSSSubjectPrefix+strings.Join([]string{fmt.Sprintf("%d", id), SSActStop}, "."),
					[]byte{},
				); err != nil {
					logger.App().Errorf("publish stop signal err: %s-%s", TGSpiderSSSubjectPrefix+SSActStop, err.Error())
					continue
				}

				if err := structure.UpdateProcess(id, 7); err != nil {
					logger.App().Errorf("update process error: %d-%d-%s", id, 7, err.Error())
					continue
				}

				logger.App().Infof("publish stop tg spider account id success : %d ", id)
			}
		}

		if done {
			break
		}
	}
}

func Handler(msg *nats2.Msg) {
	if !strings.HasPrefix(msg.Subject, TGSpiderOBESubjectPrefix) {
		logger.App().Warnf("Receive message not prefix from TGSpiderOBE: %s", msg.Subject)
		return
	}

	logger.App().Infof("Receive message by TGSpiderOBE: %s - %s", msg.Subject, string(msg.Data))

	behavior := strings.TrimPrefix(msg.Subject, TGSpiderOBESubjectPrefix)
	subs := strings.Split(behavior, ".")
	if len(subs) < 2 {
		logger.App().Errorf("Receive message subject not completed: %s", msg.Subject)
		return
	}

	id, err := strconv.ParseInt(subs[0], 10, 64)
	if err != nil {
		logger.App().Errorf("Receive message first part not id : %s", subs[0])
		return
	}
	action := subs[1]

	value, exist := _idMap.Load(fmt.Sprintf("%d", id))
	if !exist {
		logger.App().Errorf("Receive message from not exist id : %d", id)
		return
	}
	status := value.(int)

	switch action {
	case OBEActAck: // 确认
		{
			if status == 1 {
				logger.App().Infof("Receive message from already ack id: %s-%d-%d", msg.Subject, id, status)
				return
			}

			_idMap.Store(fmt.Sprintf("%d", id), 1) // 确认该id完成

			if err = structure.UpdateProcess(uint(id), 3); err != nil {
				logger.App().Errorf("update process error: %d-%d-%s", id, 3, err.Error())
				return
			}
		}
	case OBEActCode: // 要求下发验证码
		{
			if status == 0 {
				logger.App().Infof("Receive message from not ack id: %s-%d-%d", msg.Subject, id, status)
				return
			}

			if err = structure.UpdateProcess(uint(id), 4); err != nil {
				logger.App().Errorf("update process error: %d-%d-%s", id, 4, err.Error())
				return
			}
		}
	case OBEActSession: // 告知Session
		{
			if status == 0 {
				logger.App().Infof("Receive message from not ack id: %s-%d-%d", msg.Subject, id, status)
				return
			}

			if tx := mysql.Instance().Model(new(structure.TGSpiderAccount)).Where("id = ?", id).UpdateColumn("session", string(msg.Data)); tx.Error != nil {
				logger.App().Errorf("update session error: %d-%s-%s", id, string(msg.Data), tx.Error.Error())
				return
			}
		}
	case OBEActReady: // 告知已经准备好
		{
			if status == 0 {
				logger.App().Infof("Receive message from not ack id: %s-%d-%d", msg.Subject, id, status)
				return
			}

			if err = structure.UpdateProcess(uint(id), 5); err != nil {
				logger.App().Errorf("update process error: %d-%d-%s", id, 5, err.Error())
				return
			}
		}
	case OBEActOrderDone: // 告知任务已完成
		{
			if status == 0 {
				logger.App().Infof("Receive message from not ack id: %s-%d-%d", msg.Subject, id, status)
				return
			}

			// 回到已经准备好的情况等待任务下发
			if err = structure.UpdateProcess(uint(id), 5); err != nil {
				logger.App().Errorf("update process error: %d-%d-%s", id, 5, err.Error())
				return
			}
		}
	case OBEActStoped: // 告知已经停止
		{
			if status == 0 {
				logger.App().Infof("Receive message from not ack id: %s-%d-%d", msg.Subject, id, status)
				return
			}

			if err = structure.UpdateProcess(uint(id), 1); err != nil {
				logger.App().Errorf("update process error: %d-%d-%s", id, 1, err.Error())
				return
			}

			_idMap.Delete(fmt.Sprintf("%d", id))
		}
	default:
		{
			logger.App().Infof("Receive message from unknow action: %s-%s", msg.Subject, action)
		}
	}
}
