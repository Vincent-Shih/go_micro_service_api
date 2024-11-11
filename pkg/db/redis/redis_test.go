package redis_cache

import (
	"context"
	"encoding/json"
	"errors"
	"go_micro_service_api/pkg/cus_err"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisCache_Get(t *testing.T) {
	// Create mock
	client, mock := redismock.NewClientMock()
	cache := NewRedisCache(client)

	// Test cases
	tests := []struct {
		name     string
		key      string
		mockFunc func()
		want     string
		wantErr  bool
		errCode  int
	}{
		{
			name: "Successful Get",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectGet("testKey").SetVal("testValue")
			},
			want:    "testValue",
			wantErr: false,
		},
		{
			name: "Key Not Found",
			key:  "nonExistentKey",
			mockFunc: func() {
				mock.ExpectGet("nonExistentKey").RedisNil()
			},
			want:    "",
			wantErr: true,
			errCode: cus_err.ResourceNotFound,
		},
		{
			name: "UnExpected Error",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectGet("testKey").SetErr(errors.New("unexpected error"))
			},
			want:    "",
			wantErr: true,
			errCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			got, err := cache.Get(context.Background(), tt.key)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errCode, err.Code().Int())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestRedisCache_Set(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRedisCache(client)

	tests := []struct {
		name       string
		key        string
		value      string
		expiration time.Duration
		mockFunc   func()
		wantErr    bool
		errCode    int
	}{
		{
			name:       "Successful Set",
			key:        "testKey",
			value:      "testValue",
			expiration: time.Minute,
			mockFunc: func() {
				mock.ExpectSet("testKey", "testValue", time.Minute).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:       "UnExpected Error",
			key:        "testKey",
			value:      "testValue",
			expiration: time.Minute,
			mockFunc: func() {
				mock.ExpectSet("testKey", "testValue", time.Minute).SetErr(errors.New("unexpected error"))
			},
			wantErr: true,
			errCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			err := cache.Set(context.Background(), tt.key, tt.value, tt.expiration)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errCode, err.Code().Int())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestRedisCache_GetObject(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRedisCache(client)

	type testStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name     string
		key      string
		mockFunc func()
		want     testStruct
		wantErr  bool
		errCode  int
	}{
		{
			name: "Successful GetObject",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectGet("testKey").SetVal(`{"Name":"John","Age":30}`)
			},
			want:    testStruct{Name: "John", Age: 30},
			wantErr: false,
		},
		{
			name: "Key Not Found",
			key:  "nonExistentKey",
			mockFunc: func() {
				mock.ExpectGet("nonExistentKey").RedisNil()
			},
			want:    testStruct{},
			wantErr: true,
			errCode: cus_err.ResourceNotFound,
		},
		{
			name: "UnExpected Error",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectGet("testKey").SetErr(errors.New("unexpected error"))
			},
			want:    testStruct{},
			wantErr: true,
			errCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			var result testStruct
			err := cache.GetObject(context.Background(), tt.key, &result)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errCode, err.Code().Int())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, result)
			}
		})
		assert.NoError(t, mock.ExpectationsWereMet())
	}
}

func TestRedisCache_SetObject(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRedisCache(client)

	type testStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name       string
		key        string
		value      any
		expiration time.Duration
		mockFunc   func()
		wantErr    bool
		errCode    int
	}{
		{
			name:       "Successful SetObject",
			key:        "testKey",
			value:      testStruct{Name: "John", Age: 30},
			expiration: time.Minute,
			mockFunc: func() {
				expectedJSON, _ := json.Marshal(testStruct{Name: "John", Age: 30})
				mock.ExpectSet("testKey", expectedJSON, time.Minute).SetVal("OK")
			},
			wantErr: false,
		},
		{
			name:       "Marshal Error",
			key:        "testKey",
			value:      make(chan int), // Unmarshalable type
			expiration: time.Minute,
			mockFunc:   func() {}, // No mock expectation needed
			wantErr:    true,
			errCode:    cus_err.InternalServerError,
		},
		{
			name:       "Redis Set Error",
			key:        "testKey",
			value:      testStruct{Name: "John", Age: 30},
			expiration: time.Minute,
			mockFunc: func() {
				expectedJSON, _ := json.Marshal(testStruct{Name: "John", Age: 30})
				mock.ExpectSet("testKey", expectedJSON, time.Minute).SetErr(errors.New("redis error"))
			},
			wantErr: true,
			errCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			err := cache.SetObject(context.Background(), tt.key, tt.value, tt.expiration)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errCode, err.Code().Int())
			} else {
				assert.Nil(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedisCache_Delete(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRedisCache(client)

	tests := []struct {
		name     string
		keys     []string
		mockFunc func()
		wantErr  bool
		errCode  int
	}{
		{
			name: "Successful Delete",
			keys: []string{"testKey"},
			mockFunc: func() {
				mock.ExpectDel("testKey").SetVal(1)
			},
			wantErr: false,
		},
		{
			name: "Successful Delete Multiple Keys",
			keys: []string{"testKey1", "testKey2"},
			mockFunc: func() {
				mock.ExpectDel("testKey1", "testKey2").SetVal(2)
			},
			wantErr: false,
		},
		{
			name: "Key Not Found",
			keys: []string{"nonExistentKey"},
			mockFunc: func() {
				mock.ExpectDel("nonExistentKey").SetVal(0)
			},
			wantErr: true,
			errCode: cus_err.ResourceNotFound,
		},
		{
			name: "Unexpected Error",
			keys: []string{"testKey"},
			mockFunc: func() {
				mock.ExpectDel("testKey").SetErr(errors.New("unexpected error"))
			},
			wantErr: true,
			errCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			err := cache.Delete(context.Background(), tt.keys...)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errCode, err.Code().Int())
			} else {
				assert.Nil(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedisCache_GetHash(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRedisCache(client)

	type testStruct struct {
		Name string `redis:"Name"`
		Age  int    `redis:"Age"`
	}

	tests := []struct {
		name     string
		key      string
		mockFunc func()
		want     testStruct
		wantErr  bool
		errCode  int
	}{
		{
			name: "Successful",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectHGetAll("testKey").SetVal(map[string]string{"Name": "John", "Age": "30"})
			},
			want:    testStruct{Name: "John", Age: 30},
			wantErr: false,
		},
		{
			name: "Key Not Found",
			key:  "nonExistentKey",
			mockFunc: func() {
				mock.ExpectHGetAll("nonExistentKey").RedisNil()
			},
			wantErr: true,
			errCode: cus_err.ResourceNotFound,
		},
		{
			name: "Unexpected Error",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectHGetAll("testKey").SetErr(errors.New("unexpected error"))
			},
			wantErr: true,
			errCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			var data testStruct
			err := cache.GetHash(context.Background(), tt.key, &data)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errCode, err.Code().Int())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, data)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedisCache_GetHashFields(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRedisCache(client)

	tests := []struct {
		name     string
		key      string
		mockFunc func()
		want     any
		wantErr  bool
		errCode  int
	}{
		{
			name: "Successful",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectHMGet("testKey", "Name", "Age").SetVal([]interface{}{"John", "30"})
			},
			want:    []interface{}{"John", "30"},
			wantErr: false,
		},
		{
			name: "Key Not Found",
			key:  "nonExistentKey",
			mockFunc: func() {
				mock.ExpectHMGet("nonExistentKey", "Name", "Age").RedisNil()
			},
			wantErr: true,
			errCode: cus_err.ResourceNotFound,
		},
		{
			name: "Unexpected Error",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectHMGet("testKey", "Name", "Age").SetErr(errors.New("unexpected error"))
			},
			wantErr: true,
			errCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			res, err := cache.GetHashFields(context.Background(), tt.key, "Name", "Age")

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errCode, err.Code().Int())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, res)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedisCache_SetHash(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRedisCache(client)

	type testStruct struct {
		Name string
		Age  int
	}

	tests := []struct {
		name       string
		key        string
		value      testStruct
		expiration time.Duration
		mockFunc   func()
		wantErr    bool
		errCode    int
	}{
		{
			name:       "Successful",
			key:        "testKey",
			value:      testStruct{Name: "John", Age: 30},
			expiration: time.Minute,
			mockFunc: func() {
				mock.ExpectHSet("testKey", "Name", "John", "Age", 30).SetVal(0)
				mock.ExpectExpire("testKey", time.Minute).SetVal(true)
			},
			wantErr: false,
		},
		{
			name:       "Unexpected Error",
			key:        "testKey",
			value:      testStruct{Name: "John", Age: 30},
			expiration: time.Minute,
			mockFunc: func() {
				mock.ExpectHSet("testKey", "Name", "John", "Age", 30).SetErr(errors.New("unexpected error"))
			},
			wantErr: true,
			errCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			err := cache.SetHash(
				context.Background(),
				tt.expiration,
				tt.key,
				"Name", tt.value.Name, "Age", tt.value.Age,
			)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errCode, err.Code().Int())
			} else {
				assert.Nil(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRedisCache_Incr(t *testing.T) {
	client, mock := redismock.NewClientMock()
	cache := NewRedisCache(client)

	tests := []struct {
		name     string
		key      string
		mockFunc func()
		want     int
		wantErr  bool
		errCode  int
	}{
		{
			name: "Successful Incr",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectIncr("testKey").SetVal(1)
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Key Not Found",
			key:  "nonExistentKey",
			mockFunc: func() {
				mock.ExpectIncr("nonExistentKey").RedisNil()
			},
			wantErr: true,
			errCode: cus_err.ResourceNotFound,
		},
		{
			name: "Unexpected Error",
			key:  "testKey",
			mockFunc: func() {
				mock.ExpectIncr("testKey").SetErr(errors.New("unexpected error"))
			},
			wantErr: true,
			errCode: cus_err.InternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockFunc()

			count, err := cache.Incr(context.Background(), tt.key)

			if tt.wantErr {
				assert.NotNil(t, err)
				assert.Equal(t, tt.errCode, err.Code().Int())
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.want, int(count))
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
