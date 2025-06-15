package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type SettingsStorage interface {
	Get() string
	Put(value string) error
}

type memoryStorage struct {
	value string
	mu    sync.Mutex
}

func NewMemoryStorage(initialValue string) SettingsStorage {
	return &memoryStorage{value: initialValue}
}

func (s *memoryStorage) Get() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.value
}

func (s *memoryStorage) Put(value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.value = value
	return nil
}

type fileStorage struct {
	value    string
	mu       sync.Mutex
	filePath string
}

func NewFileStorage(filePath string) SettingsStorage {
	data, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("SettingsStorage: %v", err)
	}
	return &fileStorage{value: string(data), filePath: filePath}
}

func (s *fileStorage) Get() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.value
}

func (s *fileStorage) Put(value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.value = value
	return os.WriteFile(s.filePath, []byte(value), 0666)
}

const (
	s3Endpoint = "https://storage.yandexcloud.net"
	s3Region   = "ru-central1"
)

type s3Storage struct {
	value       string
	mu          sync.Mutex
	client      *s3.Client
	s3Bucket    string
	s3ObjectKey string
}

func NewS3Storage(s3AccessKey, s3Secret, s3Bucket, s3ObjectKey string) SettingsStorage {
	client := s3.New(s3.Options{
		BaseEndpoint: aws.String(s3Endpoint),
		Region:       s3Region,
		Credentials:  credentials.NewStaticCredentialsProvider(s3AccessKey, s3Secret, ""),
	})

	obj, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3ObjectKey),
	})

	if err != nil {
		var errNoSuchKey *types.NoSuchKey
		if errors.As(err, &errNoSuchKey) {
			return &s3Storage{
				value:       "",
				client:      client,
				s3Bucket:    s3Bucket,
				s3ObjectKey: s3ObjectKey,
			}
		}

		log.Fatalf("SettingsStorage: %v", err)
	}

	defer obj.Body.Close()

	data, err := io.ReadAll(obj.Body)
	if err != nil {
		log.Fatalf("SettingsStorage: %v", err)
	}

	return &s3Storage{
		value:       string(data),
		client:      client,
		s3Bucket:    s3Bucket,
		s3ObjectKey: s3ObjectKey,
	}
}

func (s *s3Storage) Get() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.value
}

func (s *s3Storage) Put(value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.value = value

	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.s3Bucket),
		Key:    aws.String(s.s3ObjectKey),
		Body:   io.NopCloser(bytes.NewBuffer([]byte(value))),
	})
	return err
}
