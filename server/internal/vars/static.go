package vars

import (
	"github.com/caarlos0/env/v11"
	"sync"
)

type StaticEnvStruct struct {
	ServerPort    int `env:"SERVER_PORT" envDefault:"8765"`     // Port exposed for the actual booking application
	ServerLogPort int `env:"SERVER_LOG_PORT" envDefault:"7777"` // Port exposed for logs to be viewed remotely

	PacketDropRate            float32 `env:"PACKET_DROP_RATE" envDefault:"0.10"`          // Rate of which packets are dropped (in and out)
	PacketReceiveTimeout      int     `env:"PACKET_TIMEOUT_RECEIVE" envDefault:"200"`     // Timeout for packets received in milliseconds
	PacketTTL                 int     `env:"PACKET_TTL" envDefault:"2000"`                // Maximum time to keep packets in history
	MessageAssemblerIntervals int     `env:"MESSAGE_ASSEMBLER_INTERVAL" envDefault:"500"` // Time between runs to request missing packets
	ResponseTTL               int     `env:"RESPONSE_TTL" envDefault:"5000"`              // Maximum time to keep messages in history
	ResponseIntervals         int     `env:"RESPONSE_INTERVAL" envDefault:"1000"`         // Time between runs to check for expired responses
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
