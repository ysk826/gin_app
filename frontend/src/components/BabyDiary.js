// frontend/src/components/BabyDiary.js
import React, { useState, useEffect } from "react";
import "./BabyDiary.css";

const BabyDiary = () => {
    // 24時間分のエントリー（各時間ごとのミルク、尿、便の状態）
    const [entries, setEntries] = useState([]);
    const [selectedDate, setSelectedDate] = useState(new Date());
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    // 日付のフォーマット関数（YYYY-MM-DD形式）
    const formatDate = (date) => {
        return date.toISOString().split("T")[0];
    };

    // APIから日記データを取得
    const fetchDiaryData = async () => {
        setLoading(true);
        setError(null);

        try {
            const response = await fetch(
                `http://localhost:8080/api/diary/full-day?date=${formatDate(
                    selectedDate
                )}`
            );

            if (!response.ok) {
                throw new Error("データの取得に失敗しました");
            }

            const data = await response.json();
            setEntries(data);
        } catch (err) {
            console.error("エラー:", err);
            setError(err.message);
        } finally {
            setLoading(false);
        }
    };

    // 日付が変わったら、データを再取得
    useEffect(() => {
        fetchDiaryData();
    }, [selectedDate]);

    // エントリーの更新を処理
    const handleEntryUpdate = async (hourIndex, field) => {
        try {
            // 現在の値の反転（トグル）
            const currentValue = entries[hourIndex][field];
            const newValue = currentValue ? 0 : 1;

            // エントリーの更新用オブジェクトの作成
            const updatedEntry = {
                ...entries[hourIndex],
                [field]: newValue,
            };

            // ローカル状態の更新（即時反映）
            const newEntries = [...entries];
            newEntries[hourIndex] = {
                ...newEntries[hourIndex],
                [field]: newValue,
            };
            setEntries(newEntries);

            // APIに更新を送信
            const response = await fetch(
                "http://localhost:8080/api/diary/time",
                {
                    method: "POST",
                    headers: {
                        "Content-Type": "application/json",
                    },
                    body: JSON.stringify(updatedEntry),
                }
            );

            if (!response.ok) {
                throw new Error("データの更新に失敗しました");
            }

            // 成功した場合はデータを再取得（必要に応じて）
            // fetchDiaryData();
        } catch (err) {
            console.error("更新エラー:", err);
            setError(err.message);
            // エラーの場合は元の状態に戻す
            fetchDiaryData();
        }
    };

    // 前日に移動
    const goToPreviousDay = () => {
        const prevDate = new Date(selectedDate);
        prevDate.setDate(prevDate.getDate() - 1);
        setSelectedDate(prevDate);
    };

    // 翌日に移動
    const goToNextDay = () => {
        const nextDate = new Date(selectedDate);
        nextDate.setDate(nextDate.getDate() + 1);
        setSelectedDate(nextDate);
    };

    // 今日に移動
    const goToToday = () => {
        setSelectedDate(new Date());
    };

    if (loading) {
        return <div className="loading">データを読み込み中...</div>;
    }

    if (error) {
        return <div className="error">{error}</div>;
    }

    return (
        <div className="baby-diary">
            <div className="date-navigation">
                <button onClick={goToPreviousDay}>前日</button>
                <h2>
                    {selectedDate.toLocaleDateString("ja-JP", {
                        year: "numeric",
                        month: "long",
                        day: "numeric",
                    })}
                </h2>
                <button onClick={goToNextDay}>翌日</button>
                <button onClick={goToToday}>今日</button>
            </div>

            <table className="diary-table">
                <thead>
                    <tr>
                        <th>時間</th>
                        <th>ミルク</th>
                        <th>尿</th>
                        <th>便</th>
                    </tr>
                </thead>
                <tbody>
                    {entries.map((entry, index) => {
                        // 時間の表示形式
                        const hour = String(index).padStart(2, "0");

                        return (
                            <tr key={`${hour}:00`}>
                                <td className="time-cell">{`${hour}:00`}</td>
                                <td
                                    className={`diary-cell ${
                                        entry.milk ? "checked" : ""
                                    }`}
                                    onClick={() =>
                                        handleEntryUpdate(index, "milk")
                                    }
                                >
                                    {entry.milk ? "✓" : ""}
                                </td>
                                <td
                                    className={`diary-cell ${
                                        entry.urine ? "checked" : ""
                                    }`}
                                    onClick={() =>
                                        handleEntryUpdate(index, "urine")
                                    }
                                >
                                    {entry.urine ? "✓" : ""}
                                </td>
                                <td
                                    className={`diary-cell ${
                                        entry.poop ? "checked" : ""
                                    }`}
                                    onClick={() =>
                                        handleEntryUpdate(index, "poop")
                                    }
                                >
                                    {entry.poop ? "✓" : ""}
                                </td>
                            </tr>
                        );
                    })}
                </tbody>
            </table>
        </div>
    );
};

export default BabyDiary;
