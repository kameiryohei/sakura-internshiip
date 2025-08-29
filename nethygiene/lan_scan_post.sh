#!/bin/bash

set -o nounset
set -o pipefail

# ---- .env 読み込み（同一ディレクトリ）----
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
set -a
source "${SCRIPT_DIR}/.env"
set +a


# ---- スクリプト用ローカルログ（./log/ 配下）----
LOCAL_LOG_DIR="${SCRIPT_DIR}/log"
mkdir -p "$LOCAL_LOG_DIR"
SUCCESS_LOG="${LOCAL_LOG_DIR}/lan_success.log"
ERROR_LOG="${LOCAL_LOG_DIR}/lan_error.log"
TIMESTAMP() { date '+%Y-%m-%d %H:%M:%S%z'; }
HOST="${NET_HOST}/status"

# ---- 入力ログファイル（.envのLOG_DIR配下）----
INPUT_LOG_FILE="/var/log/kern.log"

# tailで読む行数（必要に応じて .env で上書き可）
TAIL_LINES="${TAIL_LINES:-10}"

# ---- JSON エスケープ（最低限の \ と " を処理）----
json_escape() {
  local s="${1//\\/\\\\}"
  s="${s//\"/\\\"}"
  printf '%s' "$s"
}

# ---- 直近ログから [LAN_TCP_SYN]/[LAN_UDP] を抽出して (src_mac, src_ip) をTSVで取り出す ----
# 送信元MACは MAC=<dst>:<src>:<eth_hi>:<eth_lo> の2番目を採用
mapfile -t TSV < <(
  tail -n "$TAIL_LINES" "$INPUT_LOG_FILE" 2>/dev/null \
  | grep -E '\[LAN_TCP_SYN\]|\[LAN_UDP\]' \
  | awk '
      {
        # SRC= のIP
        if (match($0, /SRC=([0-9.]+)/, s)) {
          src_ip = s[1]
        } else {
          next
        }
        # MAC= の2番目（送信元MAC）を抽出
        # 想定: MAC=<dstMAC>:<srcMAC>:<eth_hi>:<eth_lo>
        src_mac = gensub(/^.*MAC=([0-9A-Fa-f:]{17}):([0-9A-Fa-f:]{17}):[0-9A-Fa-f]{2}:[0-9A-Fa-f]{2}.*$/, "\\2", 1, $0)
        # バリデーション
        if (src_mac ~ /^[0-9A-Fa-f]{2}(:[0-9A-Fa-f]{2}){5}$/) {
          # 小文字正規化
          for (i=1;i<=length(src_mac);i++) {
            c = substr(src_mac,i,1)
            printf "%s", tolower(c)
          }
          printf "\t%s\n", src_ip
        }
      }
    '
)

if [ "${#TSV[@]}" -eq 0 ]; then
  echo "$(TIMESTAMP) No target lines in ${INPUT_LOG_FILE}" | tee -a "$ERROR_LOG" >/dev/null
  # 空でも送る場合はここで空JSONを作って送る。送らない場合はexit 1。
  JSON_PAYLOAD='{"devices":{}}'
  # echo "$JSON_PAYLOAD"   # 標準出力
  # 送らず終了する場合は以下の行を有効化:
  # exit 1
fi

# ---- device1, device2, ... を組み立て ----
DEV_ENTRIES=()
idx=1
for line in "${TSV[@]}"; do
  IFS=$'\t' read -r mac ip <<<"$line"
  mac_e="$(json_escape "$mac")"
  ip_e="$(json_escape "$ip")"
  DEV_ENTRIES+=("$(
    printf '"device%d":{"mac":{"key":"%s"},"ip":{"key":"%s"}}' \
      "$idx" "$mac_e" "$ip_e"
  )")
  idx=$((idx+1))
done

# devices を連結
if [ "${#DEV_ENTRIES[@]}" -gt 0 ]; then
  DEVICES_JOINED="$(printf '%s,' "${DEV_ENTRIES[@]}")"
  DEVICES_JOINED="${DEVICES_JOINED%,}"   # 末尾カンマ削除
  JSON_PAYLOAD=$(printf '{"devices":{%s}}' "$DEVICES_JOINED")
else
  # TSVが空だった場合は空オブジェクト
  JSON_PAYLOAD='{"devices":{}}'
fi

# ---- まず標準出力にJSONを表示（要求どおり）----
echo "$JSON_PAYLOAD"

# ---- POST送信 ----
HTTP_CODE=$(curl -sS -o /dev/null -w '%{http_code}' \
  -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${NET_TOKEN}" \
  -d "$JSON_PAYLOAD" \
  "$HOST")

# ---- 成否ログ ----
if [[ "$HTTP_CODE" =~ ^2[0-9]{2}$ ]]; then
  echo "$(TIMESTAMP) Sent ${#TSV[@]} device(s) to ${HOST} [HTTP ${HTTP_CODE}]" | tee -a "$SUCCESS_LOG" >/dev/null
  exit 0
else
  echo "$(TIMESTAMP) Failed to send to ${HOST} [HTTP ${HTTP_CODE}] Payload=${JSON_PAYLOAD}" | tee -a "$ERROR_LOG" >/dev/null
  exit 1
fi
