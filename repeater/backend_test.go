package repeater

import (
	"testing"

	"github.com/dearcode/doodle/repeater/config"
)

func TestParseProjectID(t *testing.T) {
	config.Repeater.Server.SecretKey = "1qaz@WSX"
	ds := []struct {
		key string
		id  int64
	}{
		{"dhJgJns2tfBFvWVWUSGBfm1dsYVXAtTlye7csKmSgZo=", 1},
		{"+61FUC7/V/QxeZzpXV37e3jDOXEcAN3TXwFavJ1Ek9E=", 1234},
	}

	for _, data := range ds {
		id, err := parseProjectID(data.key)
		if err != nil {
			t.Fatalf(err.Error())
		}
		if id != data.id {
			t.Fatalf("invalid id:%v, expect:%v", id, data.id)
		}

		t.Logf("id:%v", id)
	}

}
