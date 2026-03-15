package model

type Image struct {
	ID          string `dynamodbav:"id" json:"id"`
	FileName    string `dynamodbav:"file_name" json:"file_name"`
	ContentType string `dynamodbav:"content_type" json:"content_type"`
	SizeBytes   int64  `dynamodbav:"size_bytes" json:"size_bytes"`
	S3Key       string `dynamodbav:"s3_key" json:"s3_key"`
	S3URL       string `dynamodbav:"s3_url" json:"s3_url"`
	UploadedAt  string `dynamodbav:"uploaded_at" json:"uploaded_at"`
}

type UploadResponse struct {
	Message string `json:"message"`
	Image   Image  `json:"image"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
