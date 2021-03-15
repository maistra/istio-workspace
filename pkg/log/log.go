package log

import (
	"os"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	zapr2 "github.com/go-logr/zapr"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	zapr "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// NilLog is a no-op logger.
var NilLog = zapr2.NewLogger(zap.New(zapcore.NewNopCore()))

// Log is the central logger for this program.
var Log = NilLog

// SetLogger sets the central logger to use.
func SetLogger(logger logr.Logger) {
	Log = logger
	logf.SetLogger(logger)
}

// CreateOperatorAwareLogger will set logging format to JSON when ran as operator or plain text when used as CLI.
func CreateOperatorAwareLogger(name string) logr.Logger {
	level := zap.InfoLevel
	if isDebugModeEnabled() {
		level = zap.DebugLevel
	}
	return CreateOperatorAwareLoggerWithLevel(name, level)
}

// CreateOperatorAwareLogger will set logging format to JSON when ran as operator or plain text when used as CLI.
func CreateOperatorAwareLoggerWithLevel(name string, level zapcore.Level) logr.Logger {
	var opts []zap.Option
	var enc zapcore.Encoder

	operator := isRunningAsOperator()
	sink := zapcore.AddSync(os.Stderr)

	if operator {
		enc, opts = configureOperatorLogging()
	} else {
		enc, opts = configureCliLogging()
	}

	opts = append(opts, zap.AddCaller(), zap.ErrorOutput(sink))

	encoder := &zapr.KubeAwareEncoder{Encoder: enc, Verbose: !operator}
	log := zap.New(zapcore.NewCore(encoder, sink, zap.NewAtomicLevelAt(level)))
	log = log.Named(name).WithOptions(opts...)

	return zapr2.NewLogger(log)
}

func configureCliLogging() (zapcore.Encoder, []zap.Option) {
	var enc zapcore.Encoder
	var opts []zap.Option
	encCfg := newCliEncoderConfig()
	if isDebugModeEnabled() {
		enc = zapcore.NewConsoleEncoder(encCfg)
	} else {
		enc = newFilteringEncoder(zapcore.NewConsoleEncoder(encCfg))
	}
	opts = append(opts, zap.Development(), zap.AddStacktrace(zap.ErrorLevel))
	return enc, opts
}

func configureOperatorLogging() (zapcore.Encoder, []zap.Option) {
	encCfg := zap.NewProductionEncoderConfig()
	enc := zapcore.NewJSONEncoder(encCfg)
	var opts []zap.Option
	opts = append(opts, zap.AddStacktrace(zap.WarnLevel),
		zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
		}))
	return enc, opts
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

func isDebugModeEnabled() bool {
	debug, _ := strconv.ParseBool(os.Getenv("IKE_LOG_DEBUG"))
	return debug
}
