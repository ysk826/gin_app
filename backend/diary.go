package main

import (
	"database/sql"
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

// 日記のエントリを作成するハンドラ
func createDiaryEntry(c *gin.Context) {
	var entry DiaryEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエスト形式です"})
		return
	}

	// 現在の時刻を取得
	now := time.Now()

	// データベースにエントリを挿入
	result, err := db.Exec(
		fmt.Sprintf(`INSERT INTO %s (
							time,
							milk,
							urine,
							poop,
							created_at)
					VALUES (?, ?, ?, ?, ?)`, tableDiary),
		entry.Time, entry.Milk, entry.Urine, entry.Poop, now)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データの保存に失敗しました"})
		return
	}

	// 挿入されたエントリのIDを取得
	id, _ := result.LastInsertId()
	entry.ID = int(id)
	entry.CreatedAt = now

	c.JSON(http.StatusCreated, entry)
}

// 日記のエントリを更新するハンドラ
func updateDiaryEntry(c *gin.Context) {
	id := c.Param("id")

	var entry DiaryEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエスト形式です"})
		return
	}

	// データベースにエントリを更新
	_, err := db.Exec(
		fmt.Sprintf(`UPDATE %s SET
						milk = ?,
						urine = ?,
						poop = ?
					WHERE id = ?`, tableDiary),
		entry.Milk, entry.Urine, entry.Poop, id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データの更新に失敗しました"})
		return
	}

	// 更新されたエントリを取得
	var updateEntry DiaryEntry
	err = db.QueryRow(
		fmt.Sprintf(`SELECT
						id,
						time,
						milk,
						urine,
						poop,
						created_at
					FROM %s
					WHERE id = ?`, tableDiary), id,
	).Scan(&updateEntry.ID,
		&updateEntry.Time,
		&updateEntry.Milk,
		&updateEntry.Urine,
		&updateEntry.Poop,
		&updateEntry.CreatedAt)

	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "エントリが見つかりません"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データの取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, updateEntry)
}

// 特定の日の全ての時間帯のエントリーを取得するハンドラ（存在しない時間帯も含む）
func getFullDayDiaryEntries(c *gin.Context) {
	dateStr := c.Query("date")

	var date time.Time
	var err error

	if dateStr == "" {
		// 日付の指定がない場合は、現在の日付を使用
		date = time.Now()
	} else {
		// クエリのパラメータから日付を取得
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "無効な日付形式です"})
			return
		}
	}

	// 指定された日付の始まりと終わり
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)

	// 24時間分のエントリーを準備
	entries := make([]DiaryEntry, 24)
	for i := 0; i < 24; i++ {
		entryTime := startOfDay.Add(time.Duration(i) * time.Hour)
		entries[i] = DiaryEntry{
			Time:  entryTime,
			Milk:  0,
			Urine: 0,
			Poop:  0,
		}
	}

	// データベースから日記のエントリを取得
	rows, err := db.Query(
		fmt.Sprintf(`SELECT
						id,
						time,
						milk,
						urine,
						poop,
						created_at FROM %s
					WHERE
						date(time) = date(?)
					ORDER BY time`, tableDiary), startOfDay)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データの取得に失敗しました"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var entry DiaryEntry
		var timeStr string
		var createdAtStr string

		err := rows.Scan(&entry.ID, &timeStr, &entry.Milk, &entry.Urine, &entry.Poop, &createdAtStr)
		if err != nil {
			continue
		}

		// 文字列をtime.Time型に変換
		entryTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
		if err != nil {
			continue
		}

		// 該当する時間帯のエントリーを更新
		hour := entryTime.Hour()
		if hour >= 0 && hour < 24 {
			entry.Time = entryTime
			entry.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAtStr)
			entries[hour] = entry
		}
	}
	c.JSON(http.StatusOK, entries)
}
