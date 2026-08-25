// Package logger 日志初始化：基于 Zap 的高性能结构化日志。
package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"golang-learn/internal/config"
)

// New 根据配置创建 zap.Logger。
// level: debug/info/warn/error；encoding: json(生产) / console(开发)；output: stdout / file
func New(cfg *config.LogConfig) *zap.Logger {
	// 1. 解析日志级别
	level := zapcore.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}

	// 2. encoder 配置
	encCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写 level（info/error）
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // 2026-08-15T10:00:00.000+0800
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // pkg/file.go:line
	}

	var encoder zapcore.Encoder
	if cfg.Encoding == "console" {
		encoder = zapcore.NewConsoleEncoder(encCfg)
	} else {
		encoder = zapcore.NewJSONEncoder(encCfg)
	}

	// 3. 输出目标：stdout 或文件
	var cores []zapcore.Core
	stdoutCore := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	cores = append(cores, stdoutCore)

	if cfg.Output == "file" {
		if err := os.MkdirAll("logs", 0o755); err == nil {
			fileCore := zapcore.NewCore(encoder, zapcore.AddSync(fileSync("logs/app.log")), level)
			cores = append(cores, fileCore)
		}
	}

	// 4. 组装 Logger
	zapLogger := zap.New(zapcore.NewTee(cores...), zap.AddCaller(), zap.AddCallerSkip(0))
	return zapLogger
}

// fileSync 打开日志文件（追加模式），失败时降级到 stderr
func fileSync(path string) zapcore.WriteSyncer {
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return zapcore.AddSync(os.Stderr)
	}
	return zapcore.AddSync(f)
}
