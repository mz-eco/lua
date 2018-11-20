package lua

import (
	"fmt"
	R "reflect"
	"strings"
)

type tags struct {
	option bool
	name   string
	skip   bool
}

func makeTags(field R.StructField) tags {

	var (
		x = tags{
			option: false,
			name:   field.Name,
			skip:   false,
		}
		tags = strings.Split(field.Tag.Get("lua"), ",")
	)

	for _, tag := range tags {

		if len(tag) == 0 {
			continue
		}

		kv := strings.Split(tag, "=")

		switch strings.ToLower(kv[0]) {
		case "-":
			x.skip = true
		case "option":
			x.option = true
		default:
			panic(
				fmt.Sprintf("unknonw tag %s for field %s", kv[0], field.Name))

		}
	}

	return x
}
