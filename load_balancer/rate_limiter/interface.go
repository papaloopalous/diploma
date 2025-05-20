package ratelimiter

import "context"

type BucketIface interface {
	GetTokens(userIP string) (int, error)      //получить текущее кол-во токенов пользователя
	AddToken(userIP string)                    //добавить токен и запустить тикер для добавления токенов
	RemoveToken(userIP string) error           //забрать токен
	GetMaxTokens(userIP string) (int, error)   //получить макс кол-во токенов
	GetRate(userIP string) (int, error)        //получить скорость восстановления токенов
	SetMaxTokens(userIP string, max int) error //установить макс кол-во токенов
	SetRate(userIP string, rate int) error     //установить скорость восстановления токенов
	AddUser(userIP string) error               //добавить пользователя
	StopAllTickers(ctx context.Context)        //остановить все тикеры добавления токенов
}

type BucketDB interface {
	FindOne(userIP string) (result Bucket, err error)
	InsertOne(userIP string, bucket Bucket) error
	UpdateOne(userIP string, updatedBucket Bucket) error
}
