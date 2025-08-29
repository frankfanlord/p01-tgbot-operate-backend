package cog_manage

import (
	"context"
	"fmt"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"jarvis/middleware/mq/nats"
	"operate-backend/core/structure"
	"time"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	StreamName          = "COGDistribution"
	DistributionSubject = "COG.Distri.Username"
)

var _close = make(chan struct{})

func doCogDistribution() {
	logger.App().Infoln("=============================================== start cog distribute ===============================================")
	defer logger.App().Infoln("=============================================== stop cog distribute ===============================================")

	js, err := jetstream.New(nats.Instance())
	if err != nil {
		logger.App().Errorf("new jetstream error : %s", err.Error())
		return
	}

	if err := js.DeleteStream(context.Background(), StreamName); err != nil {
		logger.App().Errorf("delete stream error : %s-%s", StreamName, err.Error())
	}

	if _, err = js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
		Name:      StreamName,
		Subjects:  []string{DistributionSubject},
		Retention: jetstream.WorkQueuePolicy,
		Storage:   jetstream.FileStorage,
		MaxAge:    time.Duration(24) * time.Hour,
		MaxMsgs:   -1,
	}); err != nil {
		logger.App().Errorf("create or update stream error : %s", err.Error())
		return
	}

	js.CreateOrUpdateConsumer(context.Background(), StreamName, jetstream.ConsumerConfig{})

	page := 0
	size := 100

	for {

		done := false
		select {
		case <-_close:
			{
				done = true
			}
		default:
			{
			}
		}
		if done {
			break
		}

		// send onces
		list, err := doQuery(page, size)
		if err != nil {
			logger.App().Errorf("doQuery error : %d - %d - %s", page, size, err.Error())
		}

		for _, item := range list {
			ack, pErr := js.Publish(context.Background(), DistributionSubject, []byte(fmt.Sprintf("https://t.me/%s", item.Username)))
			if pErr != nil {
				logger.App().Errorf("publish [%s] error : %s", item.Username, pErr.Error())
				continue
			}
			logger.App().Infof("publish [%s] success : %+v", item.Username, *(ack))
		}

		if len(list) < size {
			page = 0
		} else {
			page++
		}

		time.Sleep(time.Minute)
	}
}

func doQuery(page, size int) ([]structure.COG, error) {
	tx := mysql.Instance().Model(new(structure.COG))

	list := make([]structure.COG, 0)
	if tx = tx.Offset((page - 1) * size).Limit(size).Order("id desc").Find(&list); tx.Error != nil {
		return []structure.COG{}, tx.Error
	}

	return list[:], nil
}
