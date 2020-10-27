package logs

import "github.com/prometheus/common/log"

var Logger log.Logger

func InitLog(debug bool) {
	Logger = log.Base()
	err := Logger.SetFormat("logger:stdout?json=true")
	if err != nil{
		Logger.Fatalf("Set Log format error: %s", err.Error())
	}
	err = Logger.SetLevel("info")
	if err != nil {
		Logger.Fatal("Set Log level error.")
		return
	}
	if debug {
		err := Logger.SetLevel("debug")
		if err != nil {
			Logger.Fatal("Set Log level error.")
		}
	}
}
