package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/gzydong/go-chat/internal/pkg/encrypt"
	"github.com/redis/go-redis/v9"
)

type EmailStorage struct {
	redis *redis.Client
}

func NewEmailStorage(redis *redis.Client) *EmailStorage {
	return &EmailStorage{redis}
}

func (e *EmailStorage) Set(ctx context.Context, channel string, email string, code string, exp time.Duration) error {
	_, err := e.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Del(ctx, e.failName(channel, email))
		pipe.Set(ctx, e.name(channel, email), code, exp)
		return nil
	})
	return err
}

func (e *EmailStorage) Get(ctx context.Context, channel string, email string) (string, error) {
	return e.redis.Get(ctx, e.name(channel, email)).Result()
}

func (e *EmailStorage) Del(ctx context.Context, channel string, email string) error {
	return e.redis.Del(ctx, e.name(channel, email)).Err()
}

func (e *EmailStorage) Verify(ctx context.Context, channel string, email string, code string) bool {
	value, err := e.Get(ctx, channel, email)
	if err != nil || len(value) == 0 {
		return false
	}

	if value == code {
		return true
	}

	// 3分钟内同一个邮箱验证码错误次数超过5次，删除验证码
	num := e.redis.Incr(ctx, e.failName(channel, email)).Val()
	if num >= 5 {
		_, _ = e.redis.Pipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, e.name(channel, email))
			pipe.Del(ctx, e.failName(channel, email))
			return nil
		})
	} else if num == 1 {
		e.redis.Expire(ctx, e.failName(channel, email), 3*time.Minute)
	}

	return false
}

func (e *EmailStorage) name(channel string, email string) string {
	return fmt.Sprintf("im:auth:email:%s:%s", channel, encrypt.Md5(email))
}

func (e *EmailStorage) failName(channel string, email string) string {
	return fmt.Sprintf("im:auth:email_fail:%s:%s", channel, encrypt.Md5(email))
}
