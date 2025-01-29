package env

import "github.com/caarlos0/env/v11"

type StaticEnvStruct struct {
	ServerPort int `env:"SERVER_PORT" envDefault:"8765"`

	PacketDropRate float32 `env:"PACKET_DROP_RATE" envDefault:"0.10"`
}

var (
	staticEnv *StaticEnvStruct
)

func LoadStaticEnv() {
	staticEnv = &StaticEnvStruct{}
	if err := env.Parse(staticEnv); err != nil {
		panic(err)
	}
}

func GetStaticEnv() *StaticEnvStruct {
	if staticEnv == nil {
		LoadStaticEnv()
	}

	return staticEnv
}
