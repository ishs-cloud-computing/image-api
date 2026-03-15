# Image API

AWS S3와 DynamoDB를 활용한 이미지 업로드/조회 REST API  
Go 표준 라이브러리(`net/http`)만 사용하며, EC2IAMRole 기반 인증을 지원합니다.

## API Endpoints

| Method | Path                    | Description            |
| ------ | ----------------------- | ---------------------- |
| `GET`  | `/health`               | 헬스 체크              |
| `POST` | `/images`               | 이미지 업로드          |
| `GET`  | `/images/{id}`          | 이미지 메타데이터 조회 |
| `GET`  | `/images/{id}/download` | 이미지 파일 다운로드   |

### POST /images

`multipart/form-data`로 `image` 필드에 파일을 첨부합니다.

> 허용 형식: `image/jpeg`, `image/png`, `image/gif`, `image/webp`

```bash
curl -X POST http://localhost:8080/images \
    -F "image=@photo.jpg"
```

```json
{
  "message": "이미지 업로드 성공",
  "image": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "file_name": "photo.jpg",
    "content_type": "image/jpeg",
    "size_bytes": 204800,
    "s3_key": "images/550e8400-e29b-41d4-a716-446655440000.jpg",
    "s3_url": "https://my-bucket.s3.amazonaws.com/images/550e8400...",
    "uploaded_at": "2025-03-15T10:00:00Z"
  }
}
```

### GET /image/{id}

DynamoDB에서 이미지 메타데이터를 조회합니다.

```bash
curl http://localhost:8080/images/{id}
```

### GET /images/{id}/download

S3 에서 이미지를 스트리밍으로 다운로드합니다.

```bash
curl -o http://localhost:8080/images/{id}/download
```

## 환경변수

| Name                    | Required | Default          | Description                 |
| ----------------------- | -------- | ---------------- | --------------------------- |
| `S3_BUCKET_NAME`        | YES      | -                | S3 버킷 이름                |
| `DYNAMO_TABLE_NAME`     | YES      | -                | DynamoDB 테이블 이름        |
| `AWS_REGION`            | NO       | `ap-northeast-2` | AWS 리전                    |
| `PORT`                  | NO       | `8080`           | 서버 포트                   |
| `AWS_SECRET_NAME`       | NO       | -                | Secrets Manager 시크릿 이름 |
| `AWS_ACCESS_KEY_ID`     | NO       | -                | AWS 엑세스 키 (로컬 개발용) |
| `AWS_SECRET_ACCESS_KEY` | NO       | -                | AWS 시크릿 키 (로컬 개발용) |

자격증명 우선 순위

```
1. AWS_SECRET_NAME 있음 -> Secrets Manager에서 로드 (EC2 IAM Role 필요)
2. AWS_ACCESS_KEY_ID 있음 -> 환경변수 직접 사용
3. 둘 다 없음 -> SDK 기본 체인 (EC2 IAM Role, 인스턴스 프로파일 등)
```

> EC2 에서 실행할 경우 IAM Role만 연결하면 별도 자격 증명 설정 없이 동작합니다.

## DynamoDB 스키마

| 속성           | 타입   | 설명                  |
| -------------- | ------ | --------------------- |
| `ID`(PK)       | String | UUIDv4                |
| `file_name`    | String | 원본 파일명           |
| `content_type` | String | MIME 타입             |
| `size_bytes`   | Number | 파일 크기(bytes)      |
| `s3_key`       | String | S3 객체 키            |
| `s3_url`       | String | S3 객체 URL           |
| `uploaded_at`  | String | 업로드 시각 (RFC3339) |

## License

This project is licensed under [BSD-3-Claude](LICENSE).
