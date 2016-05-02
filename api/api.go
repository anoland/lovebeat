package api

import (
	"github.com/boivie/lovebeat/service"
	"github.com/op/go-logging"
	"time"
)

var log = logging.MustGetLogger("lovebeat")

var client service.ServiceIf

func now() int64 { return int64(time.Now().UnixNano() / 1e6) }

func Init(client_ service.ServiceIf) {
	client = client_
}
