package core

import (
	"errors"
	"operate-backend/config"
	"operate-backend/core/analysis_ik"
	"operate-backend/core/backend"
	"operate-backend/core/structure"
)

// Init 初始化
func Init() error {
	if err := structure.Init(); err != nil {
		return err
	}

	if err := backend.Init(
		config.Instance().Backend.Address,
		config.Instance().Backend.Prefix,
	); err != nil {
		return err
	}

	if err := analysis_ik.Init(
		config.Instance().IKServer.Address,
		config.Instance().IKServer.Prefix,
	); err != nil {
		return err
	}

	return nil
}

// Start 启动(N个子模块则channel宽度为N，保持异步同时预缴错误)
func Start() error {
	channel := make(chan any, 2)

	go backend.Start(channel)
	go analysis_ik.Start(channel)

	count := 2
	var err error

	for i := count; i > 0; i-- {
		v, ok := <-channel
		if !ok {
			err = errors.New("channel being closed")
			break
		}

		if v == nil {
			continue
		}

		if ve, yes := v.(error); yes {
			err = ve
			break
		}
	}

	return err
}

// Shutdown 关闭
func Shutdown() error {
	if err := backend.Shutdown(); err != nil {
		return err
	}

	if err := analysis_ik.Shutdown(); err != nil {
		return err
	}

	return nil
}

// LoadCache 加载缓存
func LoadCache(cfd string) error {
	if err := backend.LoadCache(); err != nil {
		return err
	}

	if err := analysis_ik.LoadCache(); err != nil {
		return err
	}

	return nil
}
