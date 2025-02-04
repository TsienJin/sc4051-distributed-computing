package chance

import (
	"math/rand"
	"server/internal/vars"
)

func DropPacket() bool {
	dropRate := vars.GetStaticEnv().PacketDropRate
	if rand.Float32() < dropRate {
		return true
	}
	return false
}
