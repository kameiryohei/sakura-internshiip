package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
    dbPath := os.Getenv("SQLITE_DB_PATH")
    if dbPath == "" {
        dbPath = "./data/app.db"
    }
    os.MkdirAll("./data", 0755)

    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (id INTEGER PRIMARY KEY, content TEXT)`)
    if err != nil {
        log.Fatal(err)
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        fmt.Fprintln(w, `<!DOCTYPE html><html lang='ja'><head><meta charset='utf-8'><title>ネットワーク機器監視システム</title><style>
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; 
            background: linear-gradient(135deg, #1e3c72, #2a5298); 
            margin: 0; 
            padding: 20px; 
            color: #333; 
        }
        .container { 
            max-width: 1200px; 
            margin: 0 auto; 
            background: #fff; 
            border-radius: 12px; 
            box-shadow: 0 8px 32px rgba(0,0,0,0.1); 
            padding: 32px; 
        }
        .header { 
            text-align: center; 
            margin-bottom: 32px; 
            border-bottom: 2px solid #e1e8ed; 
            padding-bottom: 24px; 
        }
        h1 { 
            color: #1e3c72; 
            font-size: 2.5rem; 
            margin: 0; 
            font-weight: 700; 
        }
        .subtitle { 
            color: #666; 
            font-size: 1rem; 
            margin-top: 8px; 
        }
        .status-bar { 
            background: linear-gradient(90deg, #28a745, #20c997); 
            color: white; 
            padding: 16px; 
            border-radius: 8px; 
            margin-bottom: 24px; 
            display: flex; 
            justify-content: space-between; 
            align-items: center; 
        }
        .status-item { 
            text-align: center; 
        }
        .status-label { 
            display: block; 
            font-size: 0.85rem; 
            opacity: 0.9; 
        }
        .status-value { 
            display: block; 
            font-size: 1.5rem; 
            font-weight: bold; 
            margin-top: 4px; 
        }
        .alert-banner { 
            background: #fff3cd; 
            border: 1px solid #ffeaa7; 
            color: #856404; 
            padding: 12px 16px; 
            border-radius: 6px; 
            margin-bottom: 24px; 
            display: flex; 
            align-items: center; 
            gap: 8px; 
        }
        .alert-icon { 
            font-size: 1.2rem; 
        }
        .devices-section { 
            margin-bottom: 32px; 
        }
        .section-title { 
            font-size: 1.4rem; 
            color: #1e3c72; 
            margin-bottom: 16px; 
            display: flex; 
            align-items: center; 
            gap: 8px; 
        }
        .devices-grid { 
            display: grid; 
            gap: 16px; 
        }
        .device-card { 
            border: 1px solid #e1e8ed; 
            border-radius: 8px; 
            padding: 20px; 
            display: grid; 
            grid-template-columns: 1fr auto; 
            gap: 16px; 
            align-items: center; 
            transition: all 0.2s ease; 
        }
        .device-card:hover { 
            box-shadow: 0 4px 12px rgba(0,0,0,0.1); 
            transform: translateY(-2px); 
        }
        .device-info h3 { 
            margin: 0 0 8px 0; 
            color: #333; 
            font-size: 1.1rem; 
        }
        .device-details { 
            font-size: 0.9rem; 
            color: #666; 
            line-height: 1.4; 
        }
        .device-status { 
            padding: 6px 12px; 
            border-radius: 20px; 
            font-size: 0.8rem; 
            font-weight: bold; 
            text-align: center; 
            min-width: 80px; 
        }
        .status-safe { 
            background: #d4edda; 
            color: #155724; 
        }
        .status-warning { 
            background: #fff3cd; 
            color: #856404; 
        }
        .status-danger { 
            background: #f8d7da; 
            color: #721c24; 
        }
        .broadcast-monitor { 
            background: #f8f9fa; 
            border: 1px solid #e1e8ed; 
            border-radius: 8px; 
            padding: 20px; 
        }
        .monitor-stats { 
            display: grid; 
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); 
            gap: 16px; 
            margin-top: 16px; 
        }
        .stat-card { 
            text-align: center; 
            padding: 16px; 
            background: white; 
            border-radius: 6px; 
            border: 1px solid #e1e8ed; 
        }
        .stat-number { 
            font-size: 2rem; 
            font-weight: bold; 
            color: #1e3c72; 
        }
        .stat-label { 
            font-size: 0.9rem; 
            color: #666; 
            margin-top: 4px; 
        }
        .refresh-btn { 
            background: linear-gradient(45deg, #1e3c72, #2a5298); 
            color: white; 
            border: none; 
            padding: 12px 24px; 
            border-radius: 6px; 
            font-size: 1rem; 
            cursor: pointer; 
            transition: all 0.2s ease; 
            margin-top: 24px; 
            display: block; 
            margin-left: auto; 
            margin-right: auto; 
        }
        .refresh-btn:hover { 
            transform: translateY(-2px); 
            box-shadow: 0 4px 12px rgba(30, 60, 114, 0.3); 
        }
        </style></head><body><div class='container'>`)
        
        fmt.Fprintln(w, `<div class='header'>`)
        fmt.Fprintln(w, `<h1>🛡️ ネットワーク機器監視システム</h1>`)
        fmt.Fprintln(w, `<div class='subtitle'>リアルタイム機器検出・脅威分析ダッシュボード</div>`)
        fmt.Fprintln(w, `</div>`)
        
        fmt.Fprintln(w, `<div class='status-bar'>`)
        fmt.Fprintln(w, `<div class='status-item'><span class='status-label'>監視状態</span><span class='status-value'>🟢 アクティブ</span></div>`)
        fmt.Fprintln(w, `<div class='status-item'><span class='status-label'>検出機器数</span><span class='status-value'>12台</span></div>`)
        fmt.Fprintln(w, `<div class='status-item'><span class='status-label'>最終更新</span><span class='status-value'>2分前</span></div>`)
        fmt.Fprintln(w, `</div>`)
        
        fmt.Fprintln(w, `<div class='alert-banner'>`)
        fmt.Fprintln(w, `<span class='alert-icon'>⚠️</span>`)
        fmt.Fprintln(w, `<span>不明な機器が2台検出されました。詳細確認が必要です。</span>`)
        fmt.Fprintln(w, `</div>`)
        
        fmt.Fprintln(w, `<div class='devices-section'>`)
        fmt.Fprintln(w, `<h2 class='section-title'>🖥️ 検出機器一覧</h2>`)
        fmt.Fprintln(w, `<div class='devices-grid'>`)
        
        // サンプル機器データの表示
        fmt.Fprintln(w, `<div class='device-card'>`)
        fmt.Fprintln(w, `<div class='device-info'><h3>📱 iPhone 14 Pro</h3><div class='device-details'>IP: 192.168.1.105<br>MAC: 00:1B:63:84:45:E6<br>メーカー: Apple Inc.</div></div>`)
        fmt.Fprintln(w, `<div class='device-status status-safe'>安全</div>`)
        fmt.Fprintln(w, `</div>`)
        
        fmt.Fprintln(w, `<div class='device-card'>`)
        fmt.Fprintln(w, `<div class='device-info'><h3>💻 MacBook Pro</h3><div class='device-details'>IP: 192.168.1.102<br>MAC: 00:16:CB:00:11:22<br>メーカー: Apple Inc.</div></div>`)
        fmt.Fprintln(w, `<div class='device-status status-safe'>安全</div>`)
        fmt.Fprintln(w, `</div>`)
        
        fmt.Fprintln(w, `<div class='device-card'>`)
        fmt.Fprintln(w, `<div class='device-info'><h3>❓ 不明な機器</h3><div class='device-details'>IP: 192.168.1.187<br>MAC: 08:00:27:12:34:56<br>メーカー: 不明</div></div>`)
        fmt.Fprintln(w, `<div class='device-status status-warning'>要注意</div>`)
        fmt.Fprintln(w, `</div>`)
        
        fmt.Fprintln(w, `<div class='device-card'>`)
        fmt.Fprintln(w, `<div class='device-info'><h3>🚨 疑わしい機器</h3><div class='device-details'>IP: 192.168.1.199<br>MAC: 00:00:00:00:00:00<br>メーカー: 不明 (偽装の疑い)</div></div>`)
        fmt.Fprintln(w, `<div class='device-status status-danger'>危険</div>`)
        fmt.Fprintln(w, `</div>`)
        
        fmt.Fprintln(w, `</div>`)
        fmt.Fprintln(w, `</div>`)
        
        fmt.Fprintln(w, `<div class='broadcast-monitor'>`)
        fmt.Fprintln(w, `<h2 class='section-title'>📡 ブロードキャスト/マルチキャスト監視</h2>`)
        fmt.Fprintln(w, `<div class='monitor-stats'>`)
        fmt.Fprintln(w, `<div class='stat-card'><div class='stat-number'>342</div><div class='stat-label'>ブロードキャスト/分</div></div>`)
        fmt.Fprintln(w, `<div class='stat-card'><div class='stat-number'>89</div><div class='stat-label'>マルチキャスト/分</div></div>`)
        fmt.Fprintln(w, `<div class='stat-card'><div class='stat-number'>🟡</div><div class='stat-label'>警告レベル</div></div>`)
        fmt.Fprintln(w, `<div class='stat-card'><div class='stat-number'>15:32</div><div class='stat-label'>最後の異常検出</div></div>`)
        fmt.Fprintln(w, `</div>`)
        fmt.Fprintln(w, `</div>`)
        
        // 既存のメッセージ一覧は隠すか、ログとして表示
        fmt.Fprintln(w, `<details style='margin-top: 32px;'><summary style='cursor: pointer; font-weight: bold; color: #666;'>📋 システムログ</summary><ul style='margin-top: 16px; max-height: 200px; overflow-y: auto; background: #f8f9fa; padding: 16px; border-radius: 6px;'>`)
        rows, err := db.Query("SELECT id, content FROM messages ORDER BY id DESC LIMIT 10")
        if err != nil {
            fmt.Fprintf(w, "<li style='color:red;'>%s</li>", err.Error())
        } else {
            defer rows.Close()
            for rows.Next() {
                var id int
                var content string
                rows.Scan(&id, &content)
                fmt.Fprintf(w, "<li style='padding: 4px 0; border-bottom: 1px solid #e1e8ed; font-size: 0.9rem;'><b>#%d</b>: %s</li>", id, content)
            }
        }
        fmt.Fprintln(w, `</ul></details>`)
        
        fmt.Fprintln(w, `<button class='refresh-btn' onclick='window.location.reload()'>🔄 リアルタイム更新</button>`)
        fmt.Fprintln(w, `</div></body></html>`)
    })

    http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
        var msg string
        if r.Method == "POST" {
            r.ParseForm()
            msg = r.FormValue("msg")
        } else {
            msg = r.URL.Query().Get("msg")
        }
        if msg == "" {
            http.Error(w, "msg required", 400)
            return
        }
        _, err := db.Exec("INSERT INTO messages(content) VALUES(?)", msg)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        http.Redirect(w, r, "/", http.StatusSeeOther)
    })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    fmt.Println("Listening on port", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
