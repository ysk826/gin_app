// frontend/src/App.js
import React from "react";
import "./App.css";
import BabyDiary from "./components/BabyDiary";

function App() {
    return (
        <div className="App">
            <header className="App-header">
                <h1>赤ちゃん日記</h1>
            </header>
            <main>
                <BabyDiary />
            </main>
            <footer>
                <p>© {new Date().getFullYear()} 赤ちゃん日記アプリ</p>
            </footer>
        </div>
    );
}

export default App;
