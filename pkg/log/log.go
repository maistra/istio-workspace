package log

import (
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	zapr2 "github.com/go-logr/zapr" //nolint:depguard //reason registers wrapper as logger

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	zapr "sigs.k8s.io/controller-runtime/pkg/log/zap" //nolint:depguard //reason registers wrapper as logger
)

// CreateOperatorAwareLogger will set logging format to JSON when ran as operator or plain text when used as CLI
func CreateOperatorAwareLogger(name string) logr.Logger {
	var opts []zap.Option
	var enc zapcore.Encoder
	var lvl zap.AtomicLevel

	operator := isRunningAsOperator()
	sink := zapcore.AddSync(os.Stderr)

	if operator {
		encCfg := zap.NewProductionEncoderConfig()
		enc = zapcore.NewJSONEncoder(encCfg)
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
		opts = append(opts, zap.AddStacktrace(zap.WarnLevel),
			zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
			}))
	} else {
		encCfg := newCliEncoderConfig()
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
		if debugLevel, found := os.LookupEnv("IKE_LOG_DEBUG"); found {
			debug, _ := strconv.ParseBool(debugLevel)
			if debug {
				zap.NewAtomicLevelAt(zap.DebugLevel)
				enc = zapcore.NewConsoleEncoder(encCfg)
			}
		} else {
			enc = newFilteringEncoder(zapcore.NewConsoleEncoder(encCfg))
		}
		opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
	}

	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))

	encoder := &zapr.KubeAwareEncoder{Encoder: enc, Verbose: !operator}
	log := zap.New(zapcore.NewCore(encoder, sink, lvl))
	log = log.Named(name).WithOptions(opts...)

	return zapr2.NewLogger(log)
}

func newCliEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string - that means it should be ignored
		MessageKey:  "msg",
		LineEnding:  zapcore.DefaultLineEnding,
		EncodeLevel: zapcore.CapitalLevelEncoder,
	}
}

func isRunningAsOperator() bool {
	_, runningInCluster := os.LookupEnv("OPERATOR_NAME")
	return runningInCluster
}
