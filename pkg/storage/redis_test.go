package storage

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/suite"
)

type RedisStorageTestSuite struct {
	suite.Suite
	mr  *miniredis.Miniredis
	rs  *RedisStorage
	ctx context.Context
}

func (s *RedisStorageTestSuite) SetupTest() {
	var err error
	s.mr, err = miniredis.Run()
	s.Require().NoError(err)

	s.rs, err = NewRedisStorage(s.mr.Addr(), "", 0)
	s.Require().NoError(err)

	s.ctx = context.Background()
}

func (s *RedisStorageTestSuite) TearDownTest() {
	s.rs.Close()
	s.mr.Close()
}

func (s *RedisStorageTestSuite) TestNewRedisStorage() {
	// Test successful connection
	rs, err := NewRedisStorage(s.mr.Addr(), "", 0)
	s.Require().NoError(err)
	s.Require().NotNil(rs)
	rs.Close()

	// Test failed connection
	addr := s.mr.Addr() // Get address before closing
	s.mr.Close()
	rs, err = NewRedisStorage(addr, "", 0)
	s.Require().Error(err)
	s.Require().Nil(rs)
}

func (s *RedisStorageTestSuite) TestGetRequestCount() {
	key := "test-key"

	// Test non-existent key
	count, err := s.rs.GetRequestCount(s.ctx, key)
	s.Require().NoError(err)
	s.Equal(int64(0), count)

	// Test existing key
	s.mr.Set(fmt.Sprintf("count:%s", key), "5")
	count, err = s.rs.GetRequestCount(s.ctx, key)
	s.Require().NoError(err)
	s.Equal(int64(5), count)
}

func (s *RedisStorageTestSuite) TestIncrementRequestCount() {
	key := "test-key"
	expiration := time.Second

	// Test first increment
	count, err := s.rs.IncrementRequestCount(s.ctx, key, expiration)
	s.Require().NoError(err)
	s.Equal(int64(1), count)

	// Test subsequent increment
	count, err = s.rs.IncrementRequestCount(s.ctx, key, expiration)
	s.Require().NoError(err)
	s.Equal(int64(2), count)

	// Verify expiration was set
	ttl := s.mr.TTL(fmt.Sprintf("count:%s", key))
	s.True(ttl > 0)
}

func (s *RedisStorageTestSuite) TestIsBlocked() {
	key := "test-key"

	// Test non-blocked key
	blocked, err := s.rs.IsBlocked(s.ctx, key)
	s.Require().NoError(err)
	s.False(blocked)

	// Test blocked key
	s.mr.Set(fmt.Sprintf("blocked:%s", key), "1")
	blocked, err = s.rs.IsBlocked(s.ctx, key)
	s.Require().NoError(err)
	s.True(blocked)
}

func (s *RedisStorageTestSuite) TestBlock() {
	key := "test-key"
	duration := time.Minute

	// Test blocking a key
	err := s.rs.Block(s.ctx, key, duration)
	s.Require().NoError(err)

	// Verify key is blocked
	blocked, err := s.rs.IsBlocked(s.ctx, key)
	s.Require().NoError(err)
	s.True(blocked)

	// Verify expiration was set
	ttl := s.mr.TTL(fmt.Sprintf("blocked:%s", key))
	s.True(ttl > 0)
}

func TestRedisStorageTestSuite(t *testing.T) {
	suite.Run(t, new(RedisStorageTestSuite))
}
