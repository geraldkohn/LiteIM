package logger

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func initLogger() {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   "./app.log", // 日志文件的位置
		MaxSize:    50,          // 在进行切割之前, 日志文件的最大大小(MB)
		MaxBackups: 10,          // 保留旧文件的最大个数
		MaxAge:     30,          // 保留旧文件的最大天数
		Compress:   false,       // 是否压缩, 归档旧文件
	}
	writeSyncer := zapcore.AddSync(lumberJackLogger)
	// encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder   // 修改时间编码器
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 在日志文件中使用大写字母记录日志级别
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	//
	core := zapcore.NewCore(encoder, writeSyncer, zapcore.DebugLevel)
	zlog := zap.New(core, zap.AddCaller()) // 将函数信息记录到日志中
	Logger = zlog.Sugar()
}

func init() {
	initLogger()
}
