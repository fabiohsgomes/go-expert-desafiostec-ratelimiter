package ratelimiter

import (
        "context"
        "testing"
        "time"

        "github.com/stretchr/testify/mock"
        "github.com/stretchr/testify/suite"
)

// MockStorage is a mock of the Storage interface using testify/mock
type MockStorage struct {
        mock.Mock
}

func (m *MockStorage) GetRequestCount(ctx context.Context, key string) (int64, error) {
        args := m.Called(ctx, key)
        return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorage) IncrementRequestCount(ctx context.Context, key string, expiration time.Duration) (int64, error) {
        args := m.Called(ctx, key, expiration)
        return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorage) IsBlocked(ctx context.Context, key string) (bool, error) {
        args := m.Called(ctx, key)
        return args.Bool(0), args.Error(1)
}

func (m *MockStorage) Block(ctx context.Context, key string, duration time.Duration) error {
        args := m.Called(ctx, key, duration)
        return args.Error(0)
}

func (m *MockStorage) Close() error {
        args := m.Called()
        return args.Error(0)
}

// RateLimiterTestSuite defines the test suite
type RateLimiterTestSuite struct {
        suite.Suite
        mockStorage *MockStorage
        ctx         context.Context
}

// SetupTest is called before each test
func (s *RateLimiterTestSuite) SetupTest() {
        s.mockStorage = new(MockStorage)
        s.ctx = context.Background()
}

// TestIPUnderLimit tests rate limiting for IPs under the limit
func (s *RateLimiterTestSuite) TestIPUnderLimit() {
        config := &Config{
                MaxRequestsPerSecond: 5,
                BlockDuration:        time.Minute,
        }
        limiter := New(s.mockStorage, config)
        key := "192.168.1.1"

        s.mockStorage.On("IsBlocked", s.ctx, key).Return(false, nil)
        s.mockStorage.On("IncrementRequestCount", s.ctx, key, time.Second).Return(int64(4), nil)

        allowed, err := limiter.IsAllowed(s.ctx, key, false)
        s.NoError(err)
        s.True(allowed)
        s.mockStorage.AssertExpectations(s.T())
}

// TestIPOverLimit tests rate limiting for IPs over the limit
func (s *RateLimiterTestSuite) TestIPOverLimit() {
        config := &Config{
                MaxRequestsPerSecond: 5,
                BlockDuration:        time.Minute,
        }
        limiter := New(s.mockStorage, config)
        key := "192.168.1.2"

        s.mockStorage.On("IsBlocked", s.ctx, key).Return(false, nil)
        s.mockStorage.On("IncrementRequestCount", s.ctx, key, time.Second).Return(int64(6), nil)
        s.mockStorage.On("Block", s.ctx, key, time.Minute).Return(nil)

        allowed, err := limiter.IsAllowed(s.ctx, key, false)
        s.NoError(err)
        s.False(allowed)
        s.mockStorage.AssertExpectations(s.T())
}

// TestTokenUnderCustomLimit tests rate limiting for tokens under their custom limit
func (s *RateLimiterTestSuite) TestTokenUnderCustomLimit() {
        config := &Config{
                MaxRequestsPerSecond: 5,
                BlockDuration:        time.Minute,
                TokenLimits: map[string]TokenConfig{
                        "test-token": {
                                MaxRequestsPerSecond: 10,
                                BlockDuration:        time.Minute,
                        },
                },
        }
        limiter := New(s.mockStorage, config)
        key := "test-token"

        s.mockStorage.On("IsBlocked", s.ctx, key).Return(false, nil)
        s.mockStorage.On("IncrementRequestCount", s.ctx, key, time.Second).Return(int64(9), nil)

        allowed, err := limiter.IsAllowed(s.ctx, key, true)
        s.NoError(err)
        s.True(allowed)
        s.mockStorage.AssertExpectations(s.T())
}

// TestTokenOverCustomLimit tests rate limiting for tokens over their custom limit
func (s *RateLimiterTestSuite) TestTokenOverCustomLimit() {
        config := &Config{
                MaxRequestsPerSecond: 5,
                BlockDuration:        time.Minute,
                TokenLimits: map[string]TokenConfig{
                        "test-token": {
                                MaxRequestsPerSecond: 10,
                                BlockDuration:        time.Minute,
                        },
                },
        }
        limiter := New(s.mockStorage, config)
        key := "test-token"

        s.mockStorage.On("IsBlocked", s.ctx, key).Return(false, nil)
        s.mockStorage.On("IncrementRequestCount", s.ctx, key, time.Second).Return(int64(11), nil)
        s.mockStorage.On("Block", s.ctx, key, time.Minute).Return(nil)

        allowed, err := limiter.IsAllowed(s.ctx, key, true)
        s.NoError(err)
        s.False(allowed)
        s.mockStorage.AssertExpectations(s.T())
}

// TestGetRemainingRequestsIP tests getting remaining requests for IP-based limiting
func (s *RateLimiterTestSuite) TestGetRemainingRequestsIP() {
        config := &Config{
                MaxRequestsPerSecond: 5,
        }
        limiter := New(s.mockStorage, config)
        key := "192.168.1.1"

        s.mockStorage.On("GetRequestCount", s.ctx, key).Return(int64(2), nil)

        remaining, err := limiter.GetRemainingRequests(s.ctx, key, false)
        s.NoError(err)
        s.Equal(3, remaining)
        s.mockStorage.AssertExpectations(s.T())
}

// TestGetRemainingRequestsToken tests getting remaining requests for token-based limiting
func (s *RateLimiterTestSuite) TestGetRemainingRequestsToken() {
        config := &Config{
                MaxRequestsPerSecond: 5,
                TokenLimits: map[string]TokenConfig{
                        "test-token": {
                                MaxRequestsPerSecond: 10,
                        },
                },
        }
        limiter := New(s.mockStorage, config)
        key := "test-token"

        s.mockStorage.On("GetRequestCount", s.ctx, key).Return(int64(4), nil)

        remaining, err := limiter.GetRemainingRequests(s.ctx, key, true)
        s.NoError(err)
        s.Equal(6, remaining)
        s.mockStorage.AssertExpectations(s.T())
}

// TestRateLimiterSuite runs the test suite
func TestRateLimiterSuite(t *testing.T) {
        suite.Run(t, new(RateLimiterTestSuite))
}
