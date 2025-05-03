package config

var BlockList map[string]bool

func InitBlockList() {
	BlockList = make(map[string]bool)
	BlockList["localhost:3000"] = true
	BlockList["admin.tidylnk.com:443"] = true
	BlockList["www.google.com:443"] = true
}
