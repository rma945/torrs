package db

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
	"torrsru/global"
	"torrsru/models/fdb"
)

var (
	mu     sync.Mutex
	isSync bool
)

func StartSync() {
	for !global.Stopped {
		syncDB()
		time.Sleep(time.Minute * time.Duration(global.DBSync))
	}
}

func syncDB() {
	slog.Info(fmt.Sprintf("Starting datbase sync from: %s", global.DBHost))

	mu.Lock()
	if isSync {
		mu.Unlock()
		return
	}
	isSync = true
	defer func() { isSync = false }()

	filetime := GetFileTime()

	mu.Unlock()
	start := time.Now()
	gcCount := 0
	for {
		ftstr := strconv.FormatInt(filetime, 10)
		slog.Info("Start fetching data")
		resp, err := http.Get(global.DBHost + "/sync/fdb/torrents?time=" + ftstr)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed connect to fdb: %s", global.DBHost), "err", err)
			slog.Info(fmt.Sprintf("Waiting %d minutes before retry...", global.DBSyncRetry))
			time.Sleep(time.Minute * time.Duration(global.DBSync))
			continue
		}

		var js *fdb.FDBRequest
		err = json.NewDecoder(resp.Body).Decode(&js)
		if err != nil {
			slog.Error("Failed to decode json", "err", err)
			os.Exit(1)
			return
		}
		resp.Body.Close()

		err = saveTorrents(js.Collections)
		if err != nil {
			slog.Error("Failed to save torrents", "err", err)
			os.Exit(1)
			return
		}

		torrents := 0
		for _, col := range js.Collections {
			if col.Value.FileTime > filetime {
				filetime = col.Value.FileTime
			}
			torrents += len(col.Value.Torrents)
		}

		err = SetFileTime(filetime)
		if err != nil {
			slog.Error("Failed to set time", "err", err)
			os.Exit(1)
			return
		}

		slog.Info(fmt.Sprintf("Saving data, found torrents: %d", torrents))

		if !js.Nextread {
			break
		}
		js = nil
		gcCount++
		if gcCount > 10 {
			runtime.GC()
			gcCount = 0
		}
	}

	slog.Info(fmt.Sprintf("End sync %s", time.Since(start)))
}

func getHash(magnet string) string {
	pos := strings.Index(magnet, "btih:")
	if pos == -1 {
		return ""
	}
	magnet = magnet[pos+5:]
	pos = strings.Index(magnet, "&")
	if pos == -1 {
		return strings.ToLower(magnet)
	}
	return strings.ToLower(magnet[:pos])
}

func ft2sec(ft int64) int64 {
	//#define TICKS_PER_SECOND 10000000
	//#define EPOCH_DIFFERENCE 11644473600LL
	return ft/10000000 - 11644473600
}
