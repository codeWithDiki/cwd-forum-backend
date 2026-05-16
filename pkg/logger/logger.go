package logger

import (
	"context"
	"gin-quickstart/pkg/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	log *zap.Logger
}

type Config struct {
	Production bool
	Level      string
}

const traceId string = "trace_id"

func NewLogger(isProduction bool, level string, isDisableStackTrace bool) *Logger {
	// Initialize logger based on debug mode and logging level from config
	var config zap.Config
	if !isProduction {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	}
	var lvl zapcore.Level
	if err := lvl.UnmarshalText([]byte(level)); err != nil {
		lvl = zapcore.DebugLevel
	}
	config.Level.SetLevel(zapcore.Level(lvl))
	// Configure encoder
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	if isDisableStackTrace {
		config.DisableStacktrace = true
	}

	// Create logger
	var err error
	log, err := config.Build(zap.AddCaller(), zap.AddCallerSkip(1))
	if err != nil {

		panic("Failed to initialize logger: " + err.Error())
	}
	return &Logger{log: log}
}

func (l *Logger) Info(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("trace_id", l.GetTraceID(ctx)))
	l.log.Info(msg, fields...)
}

func (l *Logger) Error(ctx context.Context, msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.String("trace_id", l.GetTraceID(ctx)))
	fields = append(fields, zap.Error(err))
	l.log.Error(msg, fields...)
}

func (l *Logger) Warn(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("trace_id", l.GetTraceID(ctx)))
	l.log.Warn(msg, fields...)
}

func (l *Logger) Debug(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("trace_id", l.GetTraceID(ctx)))
	l.log.Debug(msg, fields...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, fields ...zap.Field) {
	fields = append(fields, zap.String("trace_id", l.GetTraceID(ctx)))
	l.log.Fatal(msg, fields...)
}

func (l *Logger) Sync() error {
	return l.log.Sync()
}

func (l *Logger) Field(key string, value interface{}) zap.Field {
	switch v := value.(type) {
	// Primitives
	case string:
		return zap.String(key, v)
	case int:
		return zap.Int(key, v)
	case int64:
		return zap.Int64(key, v)
	case float32:
		return zap.Float32(key, v)
	case float64:
		return zap.Float64(key, v)
	case bool:
		return zap.Bool(key, v)
	case error:
		return zap.Error(v)

	// Slice / array
	case []string:
		return zap.Strings(key, v)
	case []int:
		return zap.Ints(key, v)
	case []int64:
		return zap.Int64s(key, v)
	case []float32:
		return zap.Float32s(key, v)
	case []float64:
		return zap.Float64s(key, v)
	case []bool:
		return zap.Bools(key, v)

	// Struct / object kompleks → fallback ke zap.Any
	default:
		return zap.Any(key, v)
	}
}

func (l *Logger) SetTraceID(ctx context.Context) context.Context {
	traceID, err := utils.GenUUIDV7()
	if err != nil {
		traceID = "00000000-0000-7000-8000-000000000000"
	}
	return context.WithValue(ctx, traceId, traceID)
}

func (l *Logger) GetTraceID(ctx context.Context) string {
	if gc, ok := ctx.(*gin.Context); ok && gc.Request != nil {
		ctx = gc.Request.Context()
	}
	traceID, ok := ctx.Value(traceId).(string)
	if !ok {
		return "00000000-0000-7000-8000-000000000000"
	}
	return traceID
}