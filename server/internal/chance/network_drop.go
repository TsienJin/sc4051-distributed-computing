package chance

import (
	"math/rand"
	"server/internal/env"
)

func DropPacket() bool {
	dropRate := env.GetStaticEnv().PacketDropRate
	if rand.Float32() < dropRate {
		return true
	}
	return false
}
