package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"torrsru/db"
	"torrsru/global"
	"torrsru/tgbot"
	"torrsru/web"

	"github.com/alexflint/go-arg"
)

func initLogger(level string) *slog.Logger {
	loggerLogLevel := slog.LevelInfo

	switch strings.ToLower(level) {
	case "debug":
		loggerLogLevel = slog.LevelDebug
	case "info":
		loggerLogLevel = slog.LevelInfo
	case "warn":
		loggerLogLevel = slog.LevelWarn
	case "error":
		loggerLogLevel = slog.LevelError
	default:
		loggerLogLevel = slog.LevelInfo
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: loggerLogLevel}))
}

func main() {
	var args struct {
		LogLevel     string `default:"info" arg:"--log-level,env:LOG_LEVEL" help:"log level (debug,info,warn,error)"`
		RebuildIndex bool   `default:"false" arg:"-r" help:"rebuild index and exit"`
		TMDBProxy    bool   `default:"false" arg:"--tmdb,env:TMDB_PROXY" help:"proxy for TMDB"`
		Port         string `default:"8094" arg:"-p,env:PORT" help:"port for http"`
		TGBotToken   string `default:"" arg:"--token,env:TG_TOKEN" help:"telegram bot token"`
		TGHost       string `default:"http://127.0.0.1:8081" arg:"--tgapi,env:TG_HOST" help:"telegram api host"`
		TSHost       string `default:"http://127.0.0.1:8090" arg:"--ts,env:TS_HOST" help:"TorrServer host"`
		DBPath       string `default:"" arg:"--db-path,env:DB_PATH" help:"Database path (Default to cwd)"`
		DBHost       string `default:"http://62.112.8.193:9117" arg:"--db,env:DB_HOST" help:"External sync database host"`
		DBSync       int    `default:"20" arg:"--db-sync,env:DB_SYNC" help:"External database sync delay"`
		DBSyncRetry  int    `default:"10" arg:"--db-sync-retry,env:DB_SYNC_RETRY" help:"External database sync retry"`
	}
	arg.MustParse(&args)
	slog.SetDefault(initLogger(args.LogLevel))

	if args.DBPath != "" {
		global.PWD = args.DBPath
	} else {
		pwd := filepath.Dir(os.Args[0])
		pwd, _ = filepath.Abs(pwd)
		global.PWD = pwd
	}
	slog.Info(fmt.Sprintf("Database path set to: %s", global.PWD))

	global.TMDBProxy = args.TMDBProxy
	global.TSHost = args.TSHost

	global.DBHost = args.DBHost
	global.DBSync = args.DBSync
	global.DBSyncRetry = args.DBSyncRetry

	db.Init()

	if args.RebuildIndex {
		err := db.RebuildIndex()
		if err != nil {
			slog.Error("Rebuild index error:", "err", err)
			os.Exit(1)
		} else {
			slog.Info("Rebuild index success")
		}
		return
	}

	if args.TGBotToken != "" {
		if args.TGHost == "" {
			slog.Error("Telegram host is empty. Telegram api bot need for upload 2gb files")
			os.Exit(1)
		}
		err := tgbot.Start(args.TGBotToken, args.TGHost)
		if err != nil {
			slog.Error("Start Telegram bot error:", "err", err)
			os.Exit(1)
		}
	}
	web.Start(args.Port)
}
