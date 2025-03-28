import React, { useState, useEffect } from "react";
import "./App.css";

function App() {
    const [message, setMessage] = useState("Loading...");

    useEffect(() => {
        // バックエンドAPIからデータを取得
        fetch("http://localhost:8080/api/hello")
            .then((response) => response.json())
            .then((data) => setMessage(data.message))
            .catch((error) => {
                console.error("Error fetching data:", error);
                setMessage("Error loading message");
            });
    }, []);

    return (
        <div className="App">
            <header className="App-header">
                <h1>{message}</h1>
                <p>Go + Gin + React アプリケーション</p>
            </header>
        </div>
    );
}

export default App;
