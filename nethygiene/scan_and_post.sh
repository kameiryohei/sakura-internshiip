#!/bin/bash

set -o nounset
set -o pipefail

# ---- .env 読み込み ----
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
set -a
source "${SCRIPT_DIR}/.env"
set +a

# ---- ログ設定 ----
LOG_DIR="${SCRIPT_DIR}/log"
mkdir -p "$LOG_DIR"
SUCCESS_LOG="${LOG_DIR}/success.log"
ERROR_LOG="${LOG_DIR}/error.log"
TIMESTAMP() { date '+%Y-%m-%d %H:%M:%S%z'; }

# ---- MACアドレス取得 ----
mapfile -t MACS < <(arp-scan -q -I "$IFACE" --localnet \
  | grep -oE '([0-9a-f]{2}:){5}[0-9a-f]{2}')

if [ "${#MACS[@]}" -eq 0 ]; then
  echo "$(TIMESTAMP) No MAC addresses found" | tee -a "$ERROR_LOG" >/dev/null
  exit 1
fi

# ---- JSON配列に変換----
JSON_PAYLOAD=$(printf '"%s",' "${MACS[@]}" | sed 's/,$//')
JSON_PAYLOAD="[${JSON_PAYLOAD}]"
# echo "生成されたJSON_PAYLOAD: ${JSON_PAYLOAD}"

# ---- POST送信 ----
HTTP_CODE=$(curl -sS -o /dev/null -w '%{http_code}' \
  -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer ${NET_TOKEN}" \
  -d "$JSON_PAYLOAD" \
  "$NET_HOST")

# ---- 成功判定 ----
if [[ "$HTTP_CODE" =~ ^2[0-9]{2}$ ]]; then
  echo "$(TIMESTAMP) Sent MACs (${#MACS[@]} entries) to ${NET_HOST} [HTTP ${HTTP_CODE}]" \
    | tee -a "$SUCCESS_LOG" >/dev/null
  exit 0
else
  echo "$(TIMESTAMP) Failed to send MACs to ${NET_HOST} [HTTP ${HTTP_CODE}] Payload=${JSON_PAYLOAD}" \
    | tee -a "$ERROR_LOG" >/dev/null
  exit 1
fi

