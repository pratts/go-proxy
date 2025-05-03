package blocker

import (
	"goproxy/config"
)

func ValidateIfBlocked(host string) bool {
	if _, ok := config.BlockList[host]; ok {
		return true
	}
	return false
}
