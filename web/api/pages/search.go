package pages

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"torrsru/db"

	"github.com/gin-gonic/gin"
)

func Search(c *gin.Context) {
	query := c.Query("query")
	if strings.TrimSpace(query) == "" {
		c.Status(http.StatusNoContent)
		return
	}

	trs, err := db.Search(query)
	if err != nil {
		slog.Error("Failed to get from db list", "err", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	buf, err := json.Marshal(trs)
	if err != nil {
		slog.Error("Error marshal torr list", "err", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	if gin.Mode() == gin.ReleaseMode {
		estr := query + strconv.Itoa(len(trs))
		etag := fmt.Sprintf("%x", md5.Sum([]byte(estr)))
		c.Header("ETag", etag)
		c.Header("Cache-Control", "public, max-age=3600")
	}
	c.Data(200, "application/javascript; charset=utf-8", buf)
}
