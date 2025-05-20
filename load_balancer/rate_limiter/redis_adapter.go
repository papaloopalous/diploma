package ratelimiter

import (
	"encoding/json"
	"fmt"

	"load_balancer/internal/messages"

	"github.com/go-redis/redis"
)

type RedisAdapter struct {
	Client *redis.Client
}

func (r *RedisAdapter) FindOne(userIP string) (result Bucket, err error) {
	var isEmpty bool
	err = r.Client.Watch(func(tx *redis.Tx) error {
		res, err := tx.Get(userIP).Result()
		if err == redis.Nil {
			isEmpty = true
			return nil
		}
		if err != nil {
			return err
		}

		return json.Unmarshal([]byte(res), &result)
	}, userIP)

	if err != nil {
		return result, fmt.Errorf(messages.ErrFind, userIP, err)
	}

	if isEmpty {
		return result, fmt.Errorf(messages.ErrNoData, userIP)
	}

	return result, nil
}

func (r *RedisAdapter) InsertOne(userIP string, bucket Bucket) error {
	err := r.Client.Watch(func(tx *redis.Tx) error {
		tData, err := json.Marshal(bucket)
		if err != nil {
			return err
		}

		err = r.Client.Set(userIP, tData, 0).Err()
		if err != nil {
			return err
		}

		return nil
	}, userIP)

	if err != nil {
		return fmt.Errorf(messages.ErrInsert, userIP, err)
	}

	return nil
}

func (r *RedisAdapter) UpdateOne(userIP string, updatedBucket Bucket) error {

	err := r.Client.Watch(func(tx *redis.Tx) error {
		res, err := r.Client.Get(userIP).Result()
		if err != nil && err != redis.Nil {
			return err
		}

		if err == redis.Nil {
			return fmt.Errorf(messages.ErrNoData, userIP)
		}

		var currentBucket Bucket
		err = json.Unmarshal([]byte(res), &currentBucket)
		if err != nil {
			return err
		}

		currentBucket.Rate = updatedBucket.Rate
		currentBucket.MaxTokens = updatedBucket.MaxTokens
		currentBucket.Current = updatedBucket.Current

		tData, err := json.Marshal(currentBucket)
		if err != nil {
			return err
		}

		err = r.Client.Set(userIP, tData, 0).Err()
		if err != nil {
			return err
		}

		return nil
	}, userIP)

	if err != nil {
		return fmt.Errorf(messages.ErrUpdate, userIP, err)
	}

	return nil
}
