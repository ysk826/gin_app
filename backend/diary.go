package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// 日記のエントリを表す構造体
type DiaryEntry struct {
	ID        int       `json:"id"`
	Time      time.Time `json:"time"`
	Milk      int       `json:"milk"`
	Urine     int       `json:"urine"`
	Poop      int       `json:"poop"`
	CreatedAt time.Time `json:"created_at"`
}

// 日付で日記のエントリを取得するハンドラ
func getDiaryEntriesByDate(c *gin.Context) {
	dateStr := c.Query("date")

	var date time.Time
	var err error

	// 日付の指定がない場合は、現在の日付を使用
	if dateStr == "" {
		date = time.Now()
	} else {
		// クエリのパラメータから日付を取得
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無効な日付形式です"})
			return
		}
	}

	// 日付の範囲を指定
	// 例: 2023-10-01 の場合、2023-10-01 00:00:00 から 2023-10-02 00:00:00 までの範囲
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
	endOfDay := startOfDay.Add(24 * time.Hour)

	// データベースから日記のエントリを取得
	rows, err := db.Query(
		fmt.Sprintf(`SELECT
						id,
						time,
						milk,
						urine,
						poop,
						created_at FROM %s
					WHERE time >= ?
						AND time < ?
					ORDER BY time`, tableDiary), startOfDay, endOfDay)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データの取得に失敗しました"})
		return
	}
	defer rows.Close()

	var entries []DiaryEntry
	for rows.Next() {
		var entry DiaryEntry
		// string型として渡されるため、一時変数を使用
		var timeStr string
		var createdAtStr string

		err := rows.Scan(&entry.ID, &timeStr, &entry.Milk, &entry.Urine, &entry.Poop, &createdAtStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "データの取得に失敗しました"})
			return
		}

		// 文字列をtime.Time型に変換
		entry.Time, _ = time.Parse("2006-01-02 15:04:05", timeStr)
		entry.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
		entries = append(entries, entry)
	}

	c.JSON(http.StatusOK, entries)
}
