package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	desktopConnectCodePrefix     = "desktop_connect_code:"
	desktopConnectCodeUsedPrefix = "desktop_connect_code_used:"
	desktopConnectCodeUsedMark   = "used"
)

// consumeDesktopConnectCodeScript hands a code to exactly one caller and remembers that it
// was used, so a replay is reported as "already used" rather than "expired".
var consumeDesktopConnectCodeScript = redis.NewScript(`
	local value = redis.call('GET', KEYS[1])
	if value then
		redis.call('DEL', KEYS[1])
		redis.call('SET', KEYS[2], '1', 'EX', ARGV[1])
		return value
	end
	if redis.call('EXISTS', KEYS[2]) == 1 then
		return ARGV[2]
	end
	return nil
`)

type desktopConnectCodeStore struct {
	rdb *redis.Client
}

func NewDesktopConnectCodeStore(rdb *redis.Client) service.DesktopConnectCodeStore {
	return &desktopConnectCodeStore{rdb: rdb}
}

func (s *desktopConnectCodeStore) Save(ctx context.Context, codeHash string, code service.DesktopConnectCode, ttl time.Duration) error {
	val, err := json.Marshal(code)
	if err != nil {
		return fmt.Errorf("marshal desktop connect code: %w", err)
	}
	return s.rdb.Set(ctx, desktopConnectCodePrefix+codeHash, val, ttl).Err()
}

func (s *desktopConnectCodeStore) Consume(ctx context.Context, codeHash string) (*service.DesktopConnectCode, error) {
	keys := []string{desktopConnectCodePrefix + codeHash, desktopConnectCodeUsedPrefix + codeHash}
	val, err := consumeDesktopConnectCodeScript.Run(ctx, s.rdb, keys,
		int(service.DesktopConnectCodeTTL.Seconds()), desktopConnectCodeUsedMark).Text()
	if err == redis.Nil {
		return nil, service.ErrDesktopCodeExpired
	}
	if err != nil {
		return nil, err
	}
	if val == desktopConnectCodeUsedMark {
		return nil, service.ErrDesktopCodeAlreadyUsed
	}
	var code service.DesktopConnectCode
	if err := json.Unmarshal([]byte(val), &code); err != nil {
		return nil, fmt.Errorf("unmarshal desktop connect code: %w", err)
	}
	return &code, nil
}
