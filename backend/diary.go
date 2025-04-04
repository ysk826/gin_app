package main

import (
	"database/sql"
	"fmt"
	"gin_app/config"
	"log"
	"net/http"
	"path/filepath"
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

// 日付で日記のエントリーを取得するハンドラ
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

// 日記のエントリーを作成するハンドラ
func createDiaryEntry(c *gin.Context) {
	var entry DiaryEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエスト形式です"})
		return
	}

	// 現在の時刻を取得
	now := time.Now()
	// タイムゾーンなしの形式でフォーマット
	formattedTime := entry.Time.Format(time.RFC3339)
	formattedNow := time.Now().Format(time.RFC3339)

	// データベースにエントリを挿入
	result, err := db.Exec(
		fmt.Sprintf(`INSERT INTO %s (
							time,
							milk,
							urine,
							poop,
							created_at)
					VALUES (?, ?, ?, ?, ?)`, tableDiary),
		formattedTime, entry.Milk, entry.Urine, entry.Poop, formattedNow)
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

// 日記のエントリーを更新するハンドラ
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

// 指定された時間の日記エントリーを作成または更新するハンドラ
func createOrUpdateTimeEntry(c *gin.Context) {
	var entry DiaryEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無効なリクエスト形式です"})
		return
	}

	// 指定された時間のエントリーが存在するか確認
	var existingID int
	err := db.QueryRow(
		fmt.Sprintf(`SELECT id FROM %s
					WHERE strftime('%%H:%%M', time) = strftime('%%H:%%M', ?)
					AND date(time) = date(?)`, tableDiary),
		entry.Time, entry.Time).Scan(&existingID)

	if err == sql.ErrNoRows {
		// エントリーが存在しない場合は新規作成
		now := time.Now()
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

		id, _ := result.LastInsertId()
		entry.ID = int(id)
		entry.CreatedAt = now

		c.JSON(http.StatusCreated, entry)
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データの確認に失敗しました"})
	} else {
		// エントリーが存在する場合は更新
		_, err := db.Exec(
			fmt.Sprintf(`UPDATE %s SET
							milk = ?,
							urine = ?,
							poop = ?
						WHERE id = ?`, tableDiary),
			entry.Milk, entry.Urine, entry.Poop, existingID)
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
						WHERE id = ?`, tableDiary), existingID,
		).Scan(&updateEntry.ID,
			&updateEntry.Time,
			&updateEntry.Milk,
			&updateEntry.Urine,
			&updateEntry.Poop,
			&updateEntry.CreatedAt)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "データの取得に失敗しました"})
			return
		}

		c.JSON(http.StatusOK, updateEntry)
	}
}

// 特定の日の全ての時間帯のエントリーを取得するハンドラ（存在しない時間帯も含む）
func getFullDayDiaryEntries(c *gin.Context) {
	dateStr := c.Query("date")
	log.Printf("指定された日付: %s", dateStr) // todo

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
	log.Printf("検索する日付: %s", startOfDay.Format("2006-01-02")) // todo

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
	// 日付の文字列を取得
	dateString := startOfDay.Format("2006-01-02") + "%"
	log.Printf("SQL検索パターン: %s", dateString) // todo

	// todo　削除
	// diary.goのgetFullDayDiaryEntries関数内で、クエリ部分を次のように修正
	dbPath := filepath.Join(config.Config.DbDir, config.Config.DbName)
	log.Printf("使用するデータベースファイル: %s", dbPath)
	// todo 削除
	// RFC3339形式の日付部分だけを比較
	rows, err := db.Query(
		fmt.Sprintf(`SELECT id,
							time,
							milk,
							urine,
							poop,
							created_at FROM %s
                WHERE substr(time, 1, 10) = ?`, tableDiary),
		startOfDay.Format("2006-01-02"))

	if err != nil {
		log.Printf("データベース検索エラー: %v", err) // todo
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データの取得に失敗しました"})
		return
	}
	defer rows.Close()

	// 実データの数をカウント
	dataCount := 0

	for rows.Next() {
		var entry DiaryEntry
		var timeStr string
		var createdAtStr string

		err := rows.Scan(&entry.ID, &timeStr, &entry.Milk, &entry.Urine, &entry.Poop, &createdAtStr)
		if err != nil {
			log.Printf("行のスキャンエラー: %v", err) // todo
			continue
		}

		// タイムゾーン情報を含む可能性がある時間文字列を解析
		entryTime, err := time.Parse(time.RFC3339, timeStr)
		if err != nil {
			// RFC3339形式でダメなら、他の形式も試す
			entryTime, err = time.Parse("2006-01-02T15:04:05-07:00", timeStr)
			if err != nil {
				entryTime, err = time.Parse("2006-01-02 15:04:05", timeStr)
				if err != nil {
					log.Printf("時間解析エラー: %v, timeStr: %s", err, timeStr)
					continue
				}
			}
		}

		log.Printf("データ発見: 時間=%s, id=%d, milk=%d, urine=%d, poop=%d",
			entryTime.Format("2006-01-02 15:04:05"), entry.ID, entry.Milk, entry.Urine, entry.Poop)
		dataCount++

		// 該当する時間帯のエントリーを更新
		hour := entryTime.Hour()
		if hour >= 0 && hour < 24 {
			entry.Time = entries[hour].Time // 元の時間を保持
			entry.CreatedAt, _ = time.Parse("2006-01-02 15:04:05-07:00", createdAtStr)
			entries[hour] = entry
			// todo
			log.Printf("時間 %02d:00 を更新: milk=%d, urine=%d, poop=%d",
				hour, entry.Milk, entry.Urine, entry.Poop)
		}
	}

	log.Printf("合計 %d 件のデータを見つけました", dataCount)
	log.Printf("返却するデータ数: %d", len(entries))

	// 値を持つエントリをログに出力
	for i, entry := range entries {
		if entry.Milk != 0 || entry.Urine != 0 || entry.Poop != 0 {
			log.Printf("エントリ %02d:00 - milk=%d, urine=%d, poop=%d",
				i, entry.Milk, entry.Urine, entry.Poop)
		}
	}

	c.JSON(http.StatusOK, entries)
}
