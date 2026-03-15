#!/bin/bash
# API 동작 확인용 curl 스크립트
# 사용법: ./scripts/test-api.sh

BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_IMAGE="${1:-/path/to/test.jpg}"

echo "=== 1. 헬스 체크 ==="
curl -s "$BASE_URL/health" | python3 -m json.tool

echo ""
echo "=== 2. 이미지 업로드 (POST /images) ==="
if [ ! -f "$TEST_IMAGE" ]; then
  echo "테스트 이미지가 없습니다. 1x1 PNG를 생성합니다..."
  # 최소 유효 PNG (1x1 투명 픽셀)
  printf '\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x06\x00\x00\x00\x1f\x15\xc4\x89\x00\x00\x00\nIDATx\x9cc\x00\x01\x00\x00\x05\x00\x01\r\n-\xb4\x00\x00\x00\x00IEND\xaeB`\x82' > /tmp/test.png
  TEST_IMAGE=/tmp/test.png
fi

UPLOAD_RESP=$(curl -s -X POST "$BASE_URL/images" \
  -F "image=@$TEST_IMAGE;type=image/png")
echo "$UPLOAD_RESP" | python3 -m json.tool

IMAGE_ID=$(echo "$UPLOAD_RESP" | python3 -c "import sys,json; print(json.load(sys.stdin)['image']['id'])" 2>/dev/null)

if [ -z "$IMAGE_ID" ]; then
  echo "업로드 실패. 종료합니다."
  exit 1
fi

echo ""
echo "=== 3. 메타데이터 조회 (GET /images/$IMAGE_ID) ==="
curl -s "$BASE_URL/images/$IMAGE_ID" | python3 -m json.tool

echo ""
echo "=== 4. 다운로드 (GET /images/$IMAGE_ID/download) ==="
curl -s -o /tmp/downloaded.png "$BASE_URL/images/$IMAGE_ID/download"
echo "다운로드 완료 → /tmp/downloaded.png"
ls -lh /tmp/downloaded.png

echo ""
echo "=== 5. 존재하지 않는 ID 조회 (404 확인) ==="
curl -s "$BASE_URL/images/nonexistent-id-12345" | python3 -m json.tool
