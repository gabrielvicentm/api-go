package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type PhotoStorage interface {
	UploadMotoristaPhoto(ctx context.Context, body io.Reader, originalFilename, contentType string) (string, error)
}

type ViagemDocumentStorage interface {
	UploadViagemDocument(ctx context.Context, body io.Reader, viagemID, originalFilename, contentType string) (string, error)
}

type R2Storage struct {
	bucketName    string
	publicBaseURL string
	motoristasKey string
	viagensKey    string
	uploader      *manager.Uploader
}

func NewR2StorageFromEnv(ctx context.Context) (*R2Storage, error) {
	accountID := strings.TrimSpace(os.Getenv("R2_ACCOUNT_ID"))
	accessKeyID := strings.TrimSpace(os.Getenv("R2_ACCESS_KEY_ID"))
	secretAccessKey := strings.TrimSpace(os.Getenv("R2_SECRET_ACCESS_KEY"))
	bucketName := strings.TrimSpace(os.Getenv("R2_BUCKET_NAME"))
	region := strings.TrimSpace(os.Getenv("R2_REGION"))
	endpoint := strings.TrimSpace(os.Getenv("R2_ENDPOINT"))
	publicBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("R2_PUBLIC_BASE_URL")), "/")
	motoristasPrefix := strings.Trim(strings.TrimSpace(os.Getenv("R2_MOTORISTAS_PREFIX")), "/")
	viagensPrefix := strings.Trim(strings.TrimSpace(os.Getenv("R2_VIAGENS_DOCUMENTOS_PREFIX")), "/")

	switch {
	case accessKeyID == "":
		return nil, errors.New("variavel R2_ACCESS_KEY_ID e obrigatoria")
	case secretAccessKey == "":
		return nil, errors.New("variavel R2_SECRET_ACCESS_KEY e obrigatoria")
	case bucketName == "":
		return nil, errors.New("variavel R2_BUCKET_NAME e obrigatoria")
	case publicBaseURL == "":
		return nil, errors.New("variavel R2_PUBLIC_BASE_URL e obrigatoria")
	}

	if region == "" {
		region = "auto"
	}

	if endpoint == "" {
		if accountID == "" {
			return nil, errors.New("variavel R2_ENDPOINT ou R2_ACCOUNT_ID e obrigatoria")
		}

		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	}

	if motoristasPrefix == "" {
		motoristasPrefix = "motoristas"
	}
	if viagensPrefix == "" {
		viagensPrefix = "viagens/documentos"
	}

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = true
	})

	return &R2Storage{
		bucketName:    bucketName,
		publicBaseURL: publicBaseURL,
		motoristasKey: motoristasPrefix,
		viagensKey:    viagensPrefix,
		uploader:      manager.NewUploader(client),
	}, nil
}

func (s *R2Storage) UploadMotoristaPhoto(ctx context.Context, body io.Reader, originalFilename, contentType string) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(originalFilename)))
	if ext == "" {
		ext = ".bin"
	}

	filename, err := randomFilename(ext)
	if err != nil {
		return "", err
	}

	key := strings.Trim(strings.Join([]string{s.motoristasKey, filename}, "/"), "/")
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}

	if _, err := s.uploader.Upload(ctx, input); err != nil {
		return "", err
	}

	return s.publicObjectURL(key), nil
}

func (s *R2Storage) UploadViagemDocument(ctx context.Context, body io.Reader, viagemID, originalFilename, contentType string) (string, error) {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(originalFilename)))
	if ext == "" {
		ext = ".bin"
	}

	filename, err := randomFilename(ext)
	if err != nil {
		return "", err
	}

	key := strings.Trim(strings.Join([]string{s.viagensKey, strings.TrimSpace(viagemID), filename}, "/"), "/")
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}

	if _, err := s.uploader.Upload(ctx, input); err != nil {
		return "", err
	}

	return s.publicObjectURL(key), nil
}

func (s *R2Storage) publicObjectURL(key string) string {
	baseURL := strings.TrimRight(s.publicBaseURL, "/")
	bucketPath := "/" + strings.Trim(s.bucketName, "/")

	if strings.HasSuffix(baseURL, bucketPath) {
		return baseURL + "/" + strings.TrimLeft(key, "/")
	}

	return baseURL + bucketPath + "/" + strings.TrimLeft(key, "/")
}

func randomFilename(ext string) (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(randomBytes) + ext, nil
}
