package backend

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// JSON type: represents the structure of the incoming JSON data.
type JSON struct {
	Devices map[string]struct {
		MAC struct {
			Key string `json:"key"`
		} `json:"mac"`
		IP struct {
			Key string `json:"key"`
		} `json:"ip"`
		Vendor struct {
			Key string `json:"key"`
		} `json:"vendor"`
	} `json:"devices"`
}

// StatusJSON type: represents the structure of the /status endpoint JSON data.
type StatusJSON struct {
	Devices map[string]struct {
		MAC struct {
			Key string `json:"key"`
		} `json:"mac"`
		IP struct {
			Key string `json:"key"`
		} `json:"ip"`
	} `json:"devices"`
}

// データベースインスタンス
var db *sql.DB

// SetDatabase function: データベースインスタンスを設定
func SetDatabase(database *sql.DB) {
	db = database
	log.Println("Database instance set in backend package")
}

// RunBackend function: starts the HTTP server for API endpoints.
func RunBackend() {
	// Register the handler function for the "/upload" endpoint.
	http.HandleFunc("/upload", uploadHandler)
	
	// Register the handler function for the "/status" endpoint.
	http.HandleFunc("/status", statusHandler)
	
	// ヘルスチェック用エンドポイント
	http.HandleFunc("/api/health", healthHandler)

	log.Println("Backend API endpoints registered")
	log.Println("Available endpoints:")
	log.Println("  POST /upload - デバイス情報をアップロード")
	log.Println("  POST /status - 危険機器情報をアップロード")
	log.Println("  GET /api/health - ヘルスチェック")
}

// healthHandler function: ヘルスチェック用ハンドラー
func healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== ヘルスチェック要求 ===")
	log.Printf("リクエスト元IP: %s", r.RemoteAddr)
	log.Printf("リクエスト時刻: %s", time.Now().Format("2006-01-02 15:04:05"))
	
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format("2006-01-02 15:04:05"),
		"service":   "network-monitoring-backend",
		"database":  "connected",
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	
	log.Printf("=== ヘルスチェック完了 ===\n")
}

// bearerAuth function: Bearer token認証の検証
func bearerAuth(expectedToken, authHeader string) bool {
	if authHeader == "" {
		return false
	}

	// "Bearer " プレフィックスをチェック
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	// Bearer tokenを取得
	token := strings.TrimPrefix(authHeader, "Bearer ")
	token = strings.TrimSpace(token)

	return token == expectedToken
}

// statusHandler function: handles dangerous device status updates.
func statusHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== 危険機器ステータス更新開始 ===")
	log.Printf("リクエスト時刻: %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("リクエスト元IP: %s", r.RemoteAddr)
	log.Printf("リクエストメソッド: %s", r.Method)
	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
	log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))
	
	// Ensure the request method is POST.
	if r.Method != http.MethodPost {
		log.Printf("エラー: 無効なリクエストメソッド - %s", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Bearer token認証のチェック
	authHeader := r.Header.Get("Authorization")
	log.Printf("Authorization Header: %s", authHeader)
	
	// 環境変数からトークンを取得
	expectedToken := os.Getenv("NET_TOKEN")
	
	// 環境変数が設定されていない場合のデフォルト値
	if expectedToken == "" {
		expectedToken = "default-secret-token"  // デフォルトトークン
		log.Printf("警告: NET_TOKEN環境変数が設定されていません。デフォルトトークンを使用します")
	}
	
	if !bearerAuth(expectedToken, authHeader) {
		log.Printf("エラー: Bearer token認証失敗")
		log.Printf("Expected token: [REDACTED] (length: %d)", len(expectedToken))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	log.Printf("✅ Bearer token認証成功")

	// データベース接続確認
	if db == nil {
		log.Printf("エラー: データベース接続が設定されていません")
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// JSONデータのパース
	log.Printf("危険機器JSONデータのパース開始...")
	statusData := parseStatusJSON(w, r)
	if len(statusData.Devices) == 0 {
		log.Printf("警告: 危険機器データが空です")
		return
	}

	log.Printf("受信した危険機器数: %d", len(statusData.Devices))

	// まず全ての機器を安全に設定
	_, err := db.Exec("UPDATE device SET is_dangerous = FALSE")
	if err != nil {
		log.Printf("エラー: 全機器の安全設定に失敗: %v", err)
		http.Error(w, "Database update failed", http.StatusInternalServerError)
		return
	}
	log.Printf("✅ 全機器を安全に設定しました")

	// 指定された機器を危険に設定
	dangerousCount := 0
	notFoundCount := 0
	
	for deviceKey, deviceData := range statusData.Devices {
		log.Printf("--- 危険機器処理開始 (Key: %s) ---", deviceKey)
		log.Printf("  MAC Address: %s", deviceData.MAC.Key)
		log.Printf("  IP Address: %s", deviceData.IP.Key)

		// MAC アドレスで機器を検索して危険フラグを設定
		result, err := db.Exec("UPDATE device SET is_dangerous = TRUE WHERE mac_address = ?", deviceData.MAC.Key)
		if err != nil {
			log.Printf("  ❌ 危険フラグ設定エラー: %v", err)
			continue
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("  ✅ 危険フラグ設定成功")
			dangerousCount++
		} else {
			log.Printf("  ⚠️ 機器が見つかりません (MAC: %s)", deviceData.MAC.Key)
			notFoundCount++
		}
		log.Printf("--- 危険機器処理完了 (Key: %s) ---", deviceKey)
	}

	log.Printf("危険機器処理結果サマリー:")
	log.Printf("  危険設定成功: %d件", dangerousCount)
	log.Printf("  機器未発見: %d件", notFoundCount)
	log.Printf("  処理対象: %d件", len(statusData.Devices))

	// レスポンスを返す
	response := map[string]interface{}{
		"status":           "success",
		"message":          "危険機器ステータスを正常に更新しました",
		"processed":        len(statusData.Devices),
		"dangerous_count":  dangerousCount,
		"not_found_count":  notFoundCount,
		"timestamp":        time.Now().Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("=== 危険機器ステータス更新完了 ===\n")
}

// uploadHandler function: handles device data uploads.
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("=== デバイスデータアップロード開始 ===")
	log.Printf("リクエスト時刻: %s", time.Now().Format("2006-01-02 15:04:05"))
	log.Printf("リクエスト元IP: %s", r.RemoteAddr)
	log.Printf("リクエストメソッド: %s", r.Method)
	log.Printf("Content-Type: %s", r.Header.Get("Content-Type"))
	log.Printf("User-Agent: %s", r.Header.Get("User-Agent"))
	
	// Ensure the request method is POST.
	if r.Method != http.MethodPost {
		log.Printf("エラー: 無効なリクエストメソッド - %s", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Bearer token認証のチェック
	authHeader := r.Header.Get("Authorization")
	log.Printf("Authorization Header: %s", authHeader)
	
	// 環境変数からトークンを取得
	expectedToken := os.Getenv("NET_TOKEN")
	
	// 環境変数が設定されていない場合のデフォルト値
	if expectedToken == "" {
		expectedToken = "default-secret-token"  // デフォルトトークン
		log.Printf("警告: NET_TOKEN環境変数が設定されていません。デフォルトトークンを使用します")
	}
	
	if !bearerAuth(expectedToken, authHeader) {
		log.Printf("エラー: Bearer token認証失敗")
		log.Printf("Expected token: [REDACTED] (length: %d)", len(expectedToken))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	
	log.Printf("✅ Bearer token認証成功")

	// データベース接続確認
	if db == nil {
		log.Printf("エラー: データベース接続が設定されていません")
		http.Error(w, "Database not configured", http.StatusInternalServerError)
		return
	}

	// Call the parseJSON function to handle the request.
	log.Printf("JSONデータのパース開始...")
	jsonData := parseJSON(w, r)
	if len(jsonData.Devices) == 0 {
		log.Printf("警告: デバイスデータが空です")
		return
	}

	log.Printf("受信したデバイス数: %d", len(jsonData.Devices))

	// Process and save devices to database
	successCount := 0
	errorCount := 0
	
	for deviceKey, deviceData := range jsonData.Devices {
		log.Printf("--- デバイス処理開始 (Key: %s) ---", deviceKey)
		log.Printf("  MAC Address: %s", deviceData.MAC.Key)
		log.Printf("  IP Address: %s", deviceData.IP.Key)
		log.Printf("  Vendor: %s", deviceData.Vendor.Key)

		// データベースに挿入（危険フラグは既存の値を保持）
		err := insertOrUpdateDevice(deviceData.MAC.Key, deviceData.IP.Key, deviceData.Vendor.Key)
		if err != nil {
			log.Printf("  ❌ データベース挿入エラー: %v", err)
			errorCount++
		} else {
			log.Printf("  ✅ データベース挿入成功")
			successCount++
		}
		log.Printf("--- デバイス処理完了 (Key: %s) ---", deviceKey)
	}

	log.Printf("処理結果サマリー:")
	log.Printf("  成功: %d件", successCount)
	log.Printf("  失敗: %d件", errorCount)
	log.Printf("  合計: %d件", successCount + errorCount)

	// レスポンスを返す
	response := map[string]interface{}{
		"status":       "success",
		"message":      "デバイスデータを正常に受信しました",
		"processed":    len(jsonData.Devices),
		"success_count": successCount,
		"error_count":   errorCount,
		"timestamp":    time.Now().Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("=== デバイスデータアップロード完了 ===\n")
}

// insertOrUpdateDevice function: データベースにデバイス情報を挿入または更新
func insertOrUpdateDevice(macAddress, ipAddress, vendor string) error {
	// 既存の機器かどうかをチェック
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM device WHERE mac_address = ?)", macAddress).Scan(&exists)
	if err != nil {
		return fmt.Errorf("機器存在確認エラー (MAC: %s): %v", macAddress, err)
	}

	if exists {
		// 既存機器の場合、is_dangerousを保持してIP、vendorのみ更新
		query := `UPDATE device SET ip_address = ?, vendor = ? WHERE mac_address = ?`
		_, err = db.Exec(query, ipAddress, vendor, macAddress)
	} else {
		// 新規機器の場合、is_dangerous = FALSEで挿入
		query := `INSERT INTO device (mac_address, ip_address, vendor, is_dangerous) VALUES (?, ?, ?, FALSE)`
		_, err = db.Exec(query, macAddress, ipAddress, vendor)
	}
	
	if err != nil {
		return fmt.Errorf("デバイス挿入/更新エラー (MAC: %s): %v", macAddress, err)
	}
	
	return nil
}

// parseJSON function: parses JSON requests.
func parseJSON(w http.ResponseWriter, r *http.Request) JSON {
	var data JSON
	
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("❌ JSONパースエラー: %v", err)
		http.Error(w, fmt.Sprintf("Invalid JSON data: %v", err), http.StatusBadRequest)
		return JSON{}
	}

	log.Printf("✅ JSONパース成功")
	return data
}

// parseStatusJSON function: parses status JSON requests.
func parseStatusJSON(w http.ResponseWriter, r *http.Request) StatusJSON {
	var data StatusJSON
	
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Printf("❌ 危険機器JSONパースエラー: %v", err)
		http.Error(w, fmt.Sprintf("Invalid JSON data: %v", err), http.StatusBadRequest)
		return StatusJSON{}
	}

	log.Printf("✅ 危険機器JSONパース成功")
	return data
}