package log

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

var Logger *zap.Logger

func Init() error {
	var config = zap.NewProductionConfig()
	loglevel, err := zapcore.ParseLevel(viper.GetString("log.level"))
	if err != nil {
		loglevel = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(loglevel)
	config.Encoding = "console" // json or console
	config.EncoderConfig.EncodeTime = func(time time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(time.Format("[2006-01-02 15:04:05.000]"))
	}
	config.OutputPaths = []string{"stdout"} // stdout or filename

	Logger, err = config.Build()
	if err != nil {
		return err
	}
	return nil
}
