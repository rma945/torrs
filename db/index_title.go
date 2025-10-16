package db

import (
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"torrsru/global"
	"torrsru/models/fdb"

	"github.com/blevesearch/bleve"
	bolt "go.etcd.io/bbolt"
)

var (
	indexTorrentTitle bleve.Index
)

func initIndex() error {
	mappings := bleve.NewIndexMapping()
	var err error
	indexTorrentTitle, err = bleve.Open(filepath.Join(global.PWD, "index.db"))
	if err != nil {
		indexTorrentTitle, err = bleve.NewUsing(filepath.Join(global.PWD, "index.db"), mappings, "scorch", "scorch", nil)
	}
	return err
}

func RebuildIndex() error {
	var err error
	indexTorrentTitle.Close()
	os.RemoveAll(filepath.Join(global.PWD, "index.db"))
	mappings := bleve.NewIndexMapping()
	indexTorrentTitle, err = bleve.NewUsing(filepath.Join(global.PWD, "index.db"), mappings, "scorch", "scorch", nil)
	if err != nil {
		return err
	}
	err = db.View(func(tx *bolt.Tx) error {
		torrsB := tx.Bucket([]byte("Torrents"))
		if torrsB == nil {
			return nil
		}

		batch := indexTorrentTitle.NewBatch()
		indexedTorrents := 0
		err := torrsB.ForEach(func(uniTorr, _ []byte) error {
			torrB := torrsB.Bucket(uniTorr)
			if torrB == nil {
				return fmt.Errorf("error in db struct")
			}

			title := string(torrB.Get([]byte("title")))

			if title != "" {
				if batch.Size() >= 100000 {
					err := indexTorrentTitle.Batch(batch)
					if err != nil {
						return err
					}
					indexedTorrents += batch.Size()
					slog.Info(fmt.Sprintf("Indexed torrents: %d", indexedTorrents))
					batch = indexTorrentTitle.NewBatch()
				} else {
					err := batch.Index(hex.EncodeToString(uniTorr), title)
					if err != nil {
						return err
					}
				}
			}
			return nil
		})
		if batch.Size() > 0 {
			err = indexTorrentTitle.Batch(batch)
			if err != nil {
				return err
			}
			indexedTorrents += batch.Size()
			slog.Info(fmt.Sprintf("Indexed torrents: %d", indexedTorrents))
		}
		return err
	})
	return err
}

func Search(query string) ([]*fdb.Torrent, error) {
	q := bleve.NewMatchQuery(query)
	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.Size = 1000
	searchResult, err := indexTorrentTitle.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	var list []*fdb.Torrent
	err = db.View(func(tx *bolt.Tx) error {
		torrsB := tx.Bucket([]byte("Torrents"))
		if torrsB == nil {
			return nil
		}
		for _, hit := range searchResult.Hits {
			if hit.Score > 2.0 {
				buf, err := hex.DecodeString(hit.ID)
				if err != nil {
					return err
				}
				torr, err := readTorrent(torrsB, buf)
				if err != nil {
					return err
				}
				list = append(list, torr)
			}
		}
		return nil
	})
	return list, err
}
