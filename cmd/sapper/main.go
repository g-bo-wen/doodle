package main

import (
	"flag"
	"fmt"
	"strings"

	repeaterConf "github.com/dearcode/doodle/repeater/config"
	"github.com/dearcode/doodle/util"
)

var (
	decodeServiceKey = flag.Bool("decode_service_key", false, "decode service key.")
	serviceKey       = flag.String("service_key", "", "service key.")
)

func parseServiceKey(key string) (int64, error) {
	if err := repeaterConf.Load(); err != nil {
		return 0, nil
	}

	aesKey := []byte(repeaterConf.Repeater.Server.SecretKey)
	aesKey = append(aesKey, []byte(strings.Repeat("\x00", 8-len(aesKey)%8))...)

	buf, err := util.AesDecrypt(key, aesKey)
	if err != nil {
		return 0, err
	}

	var id int64
	_, err = fmt.Sscanf(string(buf), "%x.", &id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func main() {
	flag.Parse()

	switch {
	case *decodeServiceKey:
		id, err := parseServiceKey(*serviceKey)
		if err != nil {
			panic(err)
		}

		fmt.Printf("project:%v\n", id)
	}
}
