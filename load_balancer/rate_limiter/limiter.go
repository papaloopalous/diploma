package ratelimiter

import (
	"context"
	"errors"
	"sync"
	"time"

	"load_balancer/internal/logger"
	"load_balancer/internal/messages"

	"github.com/go-redis/redis"
	"go.uber.org/zap"
)

type tokenBucket struct {
	DB               BucketDB
	tickers          map[string]*time.Ticker
	mu               sync.RWMutex
	defaultMaxTokens int
	defaultRate      int
}

type Bucket struct {
	Rate      int `json:"refillRate"`
	MaxTokens int `json:"maxTokens"`
	Current   int `json:"currentTokens"`
}

var _ BucketIface = &tokenBucket{}

func NewBucket(redisAddr string, defaultMaxTokens, defaultRate int) *tokenBucket {
	db := redis.NewClient(&redis.Options{Addr: redisAddr})
	tb := &tokenBucket{
		DB:               &RedisAdapter{Client: db},
		tickers:          make(map[string]*time.Ticker),
		defaultMaxTokens: defaultMaxTokens,
		defaultRate:      defaultRate,
	}

	go func() {
		keys, err := db.Keys("*").Result()
		if err != nil {
			logger.Log.Error(messages.ErrGetKeys, zap.Error(err))
			return
		}

		for _, key := range keys {
			tb.AddToken(key)
		}
	}()

	return tb
}

func (tb *tokenBucket) AddUser(userIP string) error {
	bckt := Bucket{
		Rate:      tb.defaultRate,
		MaxTokens: tb.defaultMaxTokens,
		Current:   1,
	}

	err := tb.DB.InsertOne(userIP, bckt)
	if err != nil {
		return err
	}

	tb.AddToken(userIP)
	return nil
}

func (tb *tokenBucket) GetTokens(userIP string) (int, error) {
	bckt, err := tb.DB.FindOne(userIP)
	if err != nil {
		return 0, err
	}

	return bckt.Current, nil
}

func (tb *tokenBucket) RemoveToken(userIP string) error {
	bckt, err := tb.DB.FindOne(userIP)
	if err != nil {
		return err
	}

	if bckt.Current <= 0 {
		return errors.New(messages.ErrNoAvailableToken)
	}

	bckt.Current--

	err = tb.DB.UpdateOne(userIP, bckt)
	if err != nil {
		return err
	}

	return nil
}

func (tb *tokenBucket) GetMaxTokens(userIP string) (int, error) {
	bckt, err := tb.DB.FindOne(userIP)
	if err != nil {
		return 0, err
	}

	return bckt.MaxTokens, nil
}

func (tb *tokenBucket) GetRate(userIP string) (int, error) {
	bckt, err := tb.DB.FindOne(userIP)
	if err != nil {
		return 0, err
	}

	return bckt.Rate, nil
}

func (tb *tokenBucket) SetMaxTokens(userIP string, max int) error {
	bckt, err := tb.DB.FindOne(userIP)
	if err != nil {
		return err
	}

	bckt.MaxTokens = max

	err = tb.DB.UpdateOne(userIP, bckt)
	if err != nil {
		return err
	}

	return nil
}

func (tb *tokenBucket) SetRate(userIP string, rate int) error {
	bckt, err := tb.DB.FindOne(userIP)
	if err != nil {
		return err
	}

	bckt.Rate = rate

	err = tb.DB.UpdateOne(userIP, bckt)
	if err != nil {
		return err
	}

	tb.AddToken(userIP)
	return nil
}

func (tb *tokenBucket) AddToken(userIP string) {
	tb.mu.RLock()
	if ticker, exists := tb.tickers[userIP]; exists {
		ticker.Stop()
	}
	tb.mu.RUnlock()

	bckt, err := tb.DB.FindOne(userIP)
	if err != nil {
		logger.Log.Error(messages.ErrAddToken, zap.Error(err))
	}

	interval := time.Second * time.Duration(bckt.Rate)
	ticker := time.NewTicker(interval)

	tb.mu.Lock()
	tb.tickers[userIP] = ticker
	tb.mu.Unlock()

	go func(userIP string) {
		for range ticker.C {
			bckt, err := tb.DB.FindOne(userIP)
			if err != nil {
				logger.Log.Error(messages.ErrAddToken, zap.Error(err))
			}

			if bckt.Current < bckt.MaxTokens {
				bckt.Current++
				logger.Log.Info(messages.InfoAddedToken, zap.String(messages.IP, userIP))
			}

			err = tb.DB.UpdateOne(userIP, bckt)
			if err != nil {
				logger.Log.Error(messages.ErrAddToken, zap.Error(err))
			}
		}
	}(userIP)
}

func (tb *tokenBucket) StopAllTickers(ctx context.Context) {
	<-ctx.Done()

	tb.mu.Lock()
	defer tb.mu.Unlock()

	for key, ticker := range tb.tickers {
		ticker.Stop()
		delete(tb.tickers, key)
	}

	logger.Log.Info(messages.InfoTickersStopped)
}
