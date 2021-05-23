#rotor

It's a rolling *io.WriteCloser* for file loggers

**Example:**
```go
package main

import (
	"github.com/marchukoff/rotor"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"os"
	"time"
)

func main() {
	rw := rotor.NewRotor(
		rotor.WithFileName("test"),
		rotor.WithFilePath("logs"),
		rotor.WithKeepFiles(2),
	)
	cfg := zap.NewProductionEncoderConfig()
	cfg.TimeKey = "time"
	cfg.EncodeTime = zapcore.RFC3339TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(cfg),
		zapcore.AddSync(io.MultiWriter(rw, zapcore.AddSync(os.Stdout))),
		zapcore.DebugLevel,
	)
	logger := zap.New(core)

	ticker := time.NewTicker(5 * time.Second)

	go func() {
		for t := range ticker.C {
			logger.Debug("now " + t.Format(time.ANSIC))
		}
	}()
	time.Sleep(20*time.Second)
	ticker.Stop()
}

```