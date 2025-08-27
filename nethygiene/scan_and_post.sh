#!/bin/bash

# arp-scan コマンドを使用して、ネットワーク上のデバイスの MAC アドレスをスキャン
MAC_ADDRESSES=$(arp-scan -q -I eth0 --localnet | grep -oE '([0-9a-f]{2}:){5}[0-9a-f]{2}')

# MAC_ADDRESSES が空でないことを確認
if [ -z "$MAC_ADDRESSES" ]; then
    echo "No MAC addresses found"
    
    exit 1
fi

# UI 側へ POST リクエストを送信
curl -X POST -H "Content-Type: application/json" -d "{\"MAC\": \"$MAC_ADDRESSES\"}" "$NET_HOST"
echo "MAC addresses sent to $NET_HOST"

exit 0