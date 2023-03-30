# rotor

It's a rolling *io.WriteCloser* for file loggers

**Example:**
```go
package main

import (
	"io"
	"os"
	"time"

	"github.com/marchukoff/rotor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	rw, err := rotor.New("test", "log", rotor.PostfixTime(time.Kitchen))
	if err != nil {
		panic(err)
	}
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "time"
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		zapcore.AddSync(io.MultiWriter(rw, zapcore.AddSync(os.Stdout))),
		zapcore.DebugLevel,
	)
	logger := zap.New(core)

	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for t := range ticker.C {
			logger.Debug("now " + t.Format(time.ANSIC))
		}
	}()
	logger.Info("start...")
	time.Sleep(2 * time.Minute)
	ticker.Stop()
	_ = rw.Close()
	logger.Warn("write to clossed writer") // redirect to stderr
}
```