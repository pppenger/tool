package source

import (
	"fmt"
	"github.com/go-redis/redis"
)

type RedisZSetScanner struct {
	Opt    *redis.Options
	Key    string
	Cursor uint64
	Match  string
	Batch  int64

	field string
	iter  *redis.ScanIterator
}

func (r *RedisZSetScanner) Open() error {
	c := redis.NewClient(r.Opt)
	if err := c.Ping().Err(); err != nil {
		return fmt.Errorf("redis ping err: %w", err)
	}

	iter := c.ZScan(r.Key, r.Cursor, r.Match, r.Batch).Iterator()
	if err := iter.Err(); err != nil {
		return fmt.Errorf("redis hscan err: %w", err)
	}

	r.iter = iter
	return nil
}

func (r *RedisZSetScanner) Next() bool {
	if !r.iter.Next() {
		return false
	}
	r.field = r.iter.Val()
	return r.iter.Next()
}

func (r *RedisZSetScanner) Record() []string {
	return []string{r.field, r.iter.Val()}
}

func (r *RedisZSetScanner) Close() error {
	return nil
}

func (r *RedisZSetScanner) Err() error {
	return r.iter.Err()
}

func FromRedisZSet(opt *redis.Options, key, match string, fun Action, options ...ScanOptions) error {
	return FromScanner(&RedisZSetScanner{
		Opt:    opt,
		Key:    key,
		Cursor: 0,
		Match:  match,
		Batch:  10,
		field:  "",
		iter:   nil,
	}, fun, options...)
}
