#!/bin/bash

set -o nounset
set -o pipefail

# ---- .env 読み込み（同一ディレクトリ）----
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
set -a
source "${SCRIPT_DIR}/.env"
set +a

# ---- ログ設定（./log/ 配下）----
LOG_DIR="${SCRIPT_DIR}/log"
mkdir -p "$LOG_DIR"
SUCCESS_LOG="${LOG_DIR}/success.log"
ERROR_LOG="${LOG_DIR}/error.log"
TIMESTAMP() { date '+%Y-%m-%d %H:%M:%S%z'; }
HOST="${NET_HOST}/upload"

# ---- JSON エスケープ関数（最低限の \ と " を処理）----
json_escape() {
  local s="${1//\\/\\\\}"
  s="${s//\"/\\\"}"
  printf '%s' "$s"
}

# ---- arp-scan 実行（IP / MAC / Vendor を行単位で取得）----
# 典型的な行例: "163.43.232.65   00:00:5e:00:01:ff       VRRP (last octet is VRID)"
mapfile -t LINES < <(
  arp-scan -I "$IFACE" --localnet 2>/dev/null \
  | grep -E '^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}[[:space:]]+[0-9a-fA-F:]{17}' \
  | awk 'NF>=2 {print $0}'
)

if [ "${#LINES[@]}" -eq 0 ]; then
  echo "$(TIMESTAMP) No devices found" | tee -a "$ERROR_LOG" >/dev/null
  exit 1
fi

# ---- 各行を device1, device2, ... に展開して JSON を組み立て（修正部分）----
DEV_ENTRIES=()
idx=1
for line in "${LINES[@]}"; do
  # 先頭2フィールド（IP, MAC）を抜き、残りを Vendor とみなす
  ip="$(awk '{print $1}' <<<"$line")"
  mac="$(awk '{print $2}' <<<"$line")"
  # 残り（3フィールド目以降）を vendor として復元
  vendor="$(sed -E "s/^[^[:space:]]+[[:space:]]+[^[:space:]]+[[:space:]]*//" <<<"$line")"
  # vendor が空なら Unknown
  if [ -z "$vendor" ]; then
    vendor="Unknown"
  fi

  # JSON に入れる前にエスケープ
  ip_e="$(json_escape "$ip")"
  mac_e="$(json_escape "$mac")"
  vendor_e="$(json_escape "$vendor")"

  # 受け取り側APIに合わせた形式に変更
  DEV_ENTRIES+=("$(
    printf '"device%d":{"mac":{"key":"%s"},"ip":{"key":"%s"},"vendor":{"key":"%s"}}' \
      "$idx" "$mac_e" "$ip_e" "$vendor_e"
  )")
  idx=$((idx+1))
done

# devices オブジェクトを連結
DEVICES_JOINED="$(printf '%s,' "${DEV_ENTRIES[@]}")"
DEVICES_JOINED="${DEVICES_JOINED%,}"   # 末尾のカンマ削除

# ---- 目的の JSON 形へ（修正部分：不要なversionフィールドを削除）----
JSON_PAYLOAD=$(printf '{"devices":{%s}}' "$DEVICES_JOINED")

# ---- POST 送信 ----
HTTP_CODE=$(curl -sS -o /dev/null -w '%{http_code}' \
  -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${NET_TOKEN}" \
  -d "$JSON_PAYLOAD" \
  "$HOST")

# ---- 成否判定・ログ ----
if [[ "$HTTP_CODE" =~ ^2[0-9]{2}$ ]]; then
  echo "$(TIMESTAMP) Sent ${#LINES[@]} device(s) to ${HOST} [HTTP ${HTTP_CODE}]" | tee -a "$SUCCESS_LOG" >/dev/null
  exit 0
else
  echo "$(TIMESTAMP) Failed to send to ${HOST} [HTTP ${HTTP_CODE}] Payload=${JSON_PAYLOAD}" | tee -a "$ERROR_LOG" >/dev/null
  exit 1
fi