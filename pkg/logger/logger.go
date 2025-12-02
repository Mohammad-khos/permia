package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a configured Zap logger
func NewLogger() *zap.Logger {
	// تنظیمات انکودر (فرمت لاگ)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // فرمت زمان خوانا
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// خروجی هم به کنسول و هم (در صورت نیاز) فایل
	consoleEncoder := zapcore.NewJSONEncoder(encoderConfig)
	
	// سطح لاگ (Debug, Info, Error)
	logLevel := zapcore.InfoLevel
	if os.Getenv("APP_ENV") == "development" {
		logLevel = zapcore.DebugLevel
	}

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), logLevel),
	)

	// ساخت لاگر با فیلدهایی مثل نام سرویس
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	
	return logger.With(zap.String("service", "core-service"))
}