package worker

import (
	"os"
	"strings"

	"github.com/anyswap/CrossChain-Bridge/log"
	"github.com/anyswap/CrossChain-Bridge/tokens"
	"github.com/fsnotify/fsnotify"
)

// AddTokenPairDynamically add token pair dynamically
func AddTokenPairDynamically() {
	pairsDir := tokens.GetTokenPairsDir()
	if pairsDir == "" {
		log.Warn("token pairs dir is empty")
		return
	}

	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("fsnotify.NewWatcher failed", "err", err)
		return
	}
	defer watch.Close()

	err = watch.Add(pairsDir)
	if err != nil {
		log.Error("watch.Add token pairs dir failed", "err", err)
		return
	}

	ops := []fsnotify.Op{
		fsnotify.Create,
		fsnotify.Write,
	}

	for {
		select {
		case ev, ok := <-watch.Events:
			if !ok {
				continue
			}
			log.Trace("fsnotify watch event", "event", ev)
			for _, op := range ops {
				if ev.Op&op == op {
					err = addTokenPair(ev.Name)
					if err != nil {
						log.Info("addTokenPair error", "configFile", ev.Name, "err", err)
					}
					break
				}
			}
		case werr, ok := <-watch.Errors:
			if !ok {
				continue
			}
			log.Warn("fsnotify watch error", "err", werr)
		}
	}
}

func addTokenPair(fileName string) error {
	if !strings.HasSuffix(fileName, ".toml") {
		return nil
	}
	fileStat, err := os.Stat(fileName)
	// ignore if file is not exist, or is directory, or is empty file
	if err != nil || fileStat.IsDir() || fileStat.Size() == 0 {
		return nil
	}
	pairConfig, err := tokens.AddPairConfig(fileName)
	if err != nil {
		return err
	}
	log.Info("addTokenPair success", "configFile", fileName, "pairID", pairConfig.PairID)
	return nil
}
