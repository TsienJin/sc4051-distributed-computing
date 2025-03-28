package vars

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"log/slog"
	"sync"
)

type StaticEnvStruct struct {
	ServerPort    int `env:"SERVER_PORT" envDefault:"8765"`     // Port exposed for the actual booking application
	ServerLogPort int `env:"SERVER_LOG_PORT" envDefault:"7777"` // Port exposed for logs to be viewed remotely

	EnableDuplicateFiltering  bool    `env:"ENABLE_DUPLICATE_FILTERING" envDefault:"true"`
	PacketDropRate            float32 `env:"PACKET_DROP_RATE" envDefault:"0.70"`         // Rate of which packets are dropped (in and out)
	PacketReceiveTimeout      int     `env:"PACKET_TIMEOUT_RECEIVE" envDefault:"20"`     // Timeout for packets received and unacked in milliseconds
	PacketTTL                 int     `env:"PACKET_TTL" envDefault:"5000"`               // Maximum time to keep packets in history
	MessageAssemblerIntervals int     `env:"MESSAGE_ASSEMBLER_INTERVAL" envDefault:"50"` // Time between runs to request missing packets
	ResponseTTL               int     `env:"RESPONSE_TTL" envDefault:"10000"`            // Maximum time to keep messages in history
	ResponseIntervals         int     `env:"RESPONSE_INTERVAL" envDefault:"500"`         // Time between runs to check for expired responses

	MatterMostWebhook string `env:"MATTERMOST_WEBHOOK" envDefault:""`
}

var (
	staticEnv *StaticEnvStruct
	onceEnv   sync.Once
)

func LoadStaticEnv() {
	staticEnv = &StaticEnvStruct{}
	if err := env.Parse(staticEnv); err != nil {
		panic(err)
	}
}

func GetStaticEnv() *StaticEnvStruct {
	onceEnv.Do(func() {
		LoadStaticEnv()
	})

	return staticEnv
}

func GetStaticEnvCopy() StaticEnvStruct {
	return *GetStaticEnv()
}

func SetEnableDuplicateFiltering(val bool) error {
	GetStaticEnv().EnableDuplicateFiltering = val
	slog.Info("[ENV] EnableDuplicateFiltering has been updated", "val", val)
	return nil
}

func SetPacketDropRate(val float32) error {
	if val < 0 || val >= 1 {
		return fmt.Errorf("val must be within bounds [0,1)")
	}

	GetStaticEnv().PacketDropRate = val
	slog.Info("[ENV] PacketDropRate has been updated", "val", val)
	return nil
}

func SetPacketReceiveTimeout(val int) error {
	if val < 0 {
		return fmt.Errorf("val must be a possitive number")
	}

	GetStaticEnv().PacketReceiveTimeout = val
	slog.Info("[ENV] PacketReceiveTimeout has been updated", "val", val)
	return nil
}

func SetPacketTTL(val int) error {
	if val < 0 {
		return fmt.Errorf("val must be a possitive number")
	}

	GetStaticEnv().PacketTTL = val
	slog.Info("[ENV] PacketTTL has been updated", "val", val)
	return nil
}

func SetMessageAssemblerIntervals(val int) error {
	if val < 0 {
		return fmt.Errorf("val must be a possitive number")
	}

	GetStaticEnv().MessageAssemblerIntervals = val
	slog.Info("[ENV] MessageAssemblerIntervals has been updated", "val", val)
	return nil
}

func SetResponseTTL(val int) error {
	if val < 0 {
		return fmt.Errorf("val must be a possitive number")
	}

	GetStaticEnv().ResponseTTL = val
	slog.Info("[ENV] ResponseTTL has been updated", "val", val)
	return nil
}

func SetResponseIntervals(val int) error {
	if val < 0 {
		return fmt.Errorf("val must be a possitive number")
	}

	GetStaticEnv().ResponseIntervals = val
	slog.Info("[ENV] ResponseIntervals has been updated", "val", val)
	return nil
}
