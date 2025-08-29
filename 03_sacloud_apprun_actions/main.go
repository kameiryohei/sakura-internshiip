package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ippanpeople/sample-go/backend"
	_ "github.com/mattn/go-sqlite3"
)

var globalDB *sql.DB

func main() {
    dbPath := os.Getenv("SQLITE_DB_PATH")
    if dbPath == "" {
        dbPath = "./data/app.db"
    }
    os.MkdirAll("./data", 0755)

    var err error
    globalDB, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatal(err)
    }
    defer globalDB.Close()

    _, err = globalDB.Exec(`CREATE TABLE IF NOT EXISTS device (
        mac_address VARCHAR(50) PRIMARY KEY,
        ip_address VARCHAR(50),
        vendor VARCHAR(50),
        is_dangerous BOOLEAN DEFAULT FALSE
    )`)
    if err != nil {
        log.Fatal(err)
    }

    // 既存テーブルにis_dangerousカラムが存在しない場合は追加
    _, err = globalDB.Exec(`ALTER TABLE device ADD COLUMN is_dangerous BOOLEAN DEFAULT FALSE`)
    if err != nil {
        // カラムが既に存在する場合はエラーを無視
        log.Printf("Column is_dangerous might already exist: %v", err)
    }

    backend.SetDatabase(globalDB)

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        fmt.Fprintln(w, `<!DOCTYPE html><html lang='ja'><head><meta charset='utf-8'><title>NetHygiene</title><style>
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
        .alert-banner.danger { 
            background: #f8d7da; 
            border: 1px solid #f5c6cb; 
            color: #721c24; 
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
        .status-danger { 
            background: #f8d7da; 
            color: #721c24; 
        }
        </style>
        <script>
        // 60秒ごとにページを自動更新（機器からの送信が1分間隔のため）
        setTimeout(function() {
            location.reload();
        }, 60000);
        
        // リアルタイム更新表示用
        let lastUpdate = new Date();
        function updateTimestamp() {
            const now = new Date();
            const diffSeconds = Math.floor((now - lastUpdate) / 1000);
            const timestampEl = document.getElementById('last-update');
            if (timestampEl) {
                if (diffSeconds < 60) {
                    timestampEl.textContent = diffSeconds + '秒前';
                } else {
                    const diffMinutes = Math.floor(diffSeconds / 60);
                    timestampEl.textContent = diffMinutes + '分前';
                }
            }
        }
        
        // 1秒ごとにタイムスタンプを更新
        setInterval(updateTimestamp, 1000);
        
        // ページロード時にタイムスタンプを初期化
        window.onload = function() {
            updateTimestamp();
        }
        </script>
        </head><body><div class='container'>`)
        
        fmt.Fprintln(w, `<div class='header'>`)
        fmt.Fprintln(w, `<h1>🛡️ NetHygiene</h1>`)
        fmt.Fprintln(w, `<div class='subtitle'>リアルタイム機器検出・脅威分析ダッシュボード</div>`)
        fmt.Fprintln(w, `</div>`)
        
        // デバイス統計の取得
        var totalDevices, dangerousDevices int
        globalDB.QueryRow("SELECT COUNT(*) FROM device").Scan(&totalDevices)
        globalDB.QueryRow("SELECT COUNT(*) FROM device WHERE is_dangerous = TRUE").Scan(&dangerousDevices)
        
        fmt.Fprintln(w, `<div class='status-bar'>`)
        fmt.Fprintln(w, `<div class='status-item'><span class='status-label'>監視状態</span><span class='status-value'>🟢 アクティブ</span></div>`)
        fmt.Fprintf(w, `<div class='status-item'><span class='status-label'>検出機器数</span><span class='status-value'>%d台</span></div>`, totalDevices)
        fmt.Fprintf(w, `<div class='status-item'><span class='status-label'>危険機器数</span><span class='status-value'>%d台</span></div>`, dangerousDevices)
        fmt.Fprintf(w, `<div class='status-item'><span class='status-label'>最終更新</span><span class='status-value' id='last-update'>更新中...</span></div>`)
        fmt.Fprintln(w, `</div>`)
        
        // アラートバナーの表示
        if dangerousDevices > 0 {
            fmt.Fprintln(w, `<div class='alert-banner danger'>`)
            fmt.Fprintln(w, `<span class='alert-icon'>🚨</span>`)
            fmt.Fprintf(w, `<span>危険機器が%d台検出されました。至急対応が必要です。</span>`, dangerousDevices)
            fmt.Fprintln(w, `</div>`)
        } else {
            fmt.Fprintln(w, `<div class='alert-banner'>`)
            fmt.Fprintln(w, `<span class='alert-icon'>✅</span>`)
            fmt.Fprintln(w, `<span>すべての機器は安全です。</span>`)
            fmt.Fprintln(w, `</div>`)
        }
        
        fmt.Fprintln(w, `<div class='devices-section'>`)
        fmt.Fprintln(w, `<h2 class='section-title'>🖥️ 検出機器一覧</h2>`)
        fmt.Fprintln(w, `<div class='devices-grid'>`)
        
        rows, err := globalDB.Query("SELECT mac_address, ip_address, vendor, is_dangerous FROM device ORDER BY is_dangerous DESC, mac_address")
        if err != nil {
            fmt.Fprintf(w, "<div class='device-card'><div class='device-info'><h3>❌ エラー</h3><div class='device-details'>%s</div></div></div>", err.Error())
        } else {
            defer rows.Close()
            deviceCount := 0
            for rows.Next() {
                var macAddress, ipAddress, vendor string
                var isDangerous bool
                rows.Scan(&macAddress, &ipAddress, &vendor, &isDangerous)
                deviceCount++
                
                // ステータス判定
                statusText := "安全"
                statusClass := "status-safe"
                
                if isDangerous {
                    statusText = "危険"
                    statusClass = "status-danger"
                }
                
                // ベンダー情報の表示調整
                vendorDisplay := vendor
                if vendor == "" {
                    vendorDisplay = "不明"
                }
                
                fmt.Fprintln(w, `<div class='device-card'>`)
                fmt.Fprintf(w, `<div class='device-info'><h3>🖥️ 機器 #%d</h3><div class='device-details'>IP: %s<br>MAC: %s<br>ベンダー: %s</div></div>`, deviceCount, ipAddress, macAddress, vendorDisplay)
                fmt.Fprintf(w, `<div class='device-status %s'>%s</div>`, statusClass, statusText)
                fmt.Fprintln(w, `</div>`)
            }
            
            // 機器が見つからない場合
            if deviceCount == 0 {
                fmt.Fprintln(w, `<div class='device-card'><div class='device-info'><h3>📭 機器なし</h3><div class='device-details'>検出された機器がありません</div></div><div class='device-status status-safe'>-</div></div>`)
            }
        }
        
        fmt.Fprintln(w, `</div>`)
        fmt.Fprintln(w, `</div>`)
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
        _, err := globalDB.Exec("INSERT INTO messages(content) VALUES(?)", msg)
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
    go backend.RunBackend()
    log.Fatal(http.ListenAndServe(":"+port, nil))
}