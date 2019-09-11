package log

import (
	"os"
	"time"

	"github.com/go-logr/logr"
	zapr2 "github.com/go-logr/zapr"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	zapr "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func CreateClusterAwareLogger() logr.Logger {
	var opts []zap.Option
	var enc zapcore.Encoder
	var lvl zap.AtomicLevel

	notInCluster := !isRunningInK8sCluster()
	sink := zapcore.AddSync(os.Stderr)

	if notInCluster {
		encCfg := newCliEncoderConfig()
		enc = zapcore.NewConsoleEncoder(encCfg)
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
		opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
	} else {
		encCfg := zap.NewProductionEncoderConfig()
		enc = zapcore.NewJSONEncoder(encCfg)
		lvl = zap.NewAtomicLevelAt(zap.InfoLevel)
		opts = append(opts, zap.AddStacktrace(zap.WarnLevel),
			zap.WrapCore(func(core zapcore.Core) zapcore.Core {
				return zapcore.NewSampler(core, time.Second, 100, 100)
			}))
	}

	opts = append(opts, zap.AddCallerSkip(1), zap.ErrorOutput(sink))

	encoder := &zapr.KubeAwareEncoder{Encoder: enc, Verbose: notInCluster}
	log := zap.New(zapcore.NewCore(encoder, sink, lvl))
	log = log.WithOptions(opts...)

	return zapr2.NewLogger(log)
}

func newCliEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string - that means it should be ignored
		MessageKey:  "M",
		LineEnding:  zapcore.DefaultLineEnding,
		EncodeLevel: zapcore.CapitalLevelEncoder,
	}
}

func isRunningInK8sCluster() bool {
	_, runningInCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	return runningInCluster
}
