package backend

import "net/http/httputil"

type BackendIface interface {
	AddConn()                         //увеличивает счётчик активных соединений
	RemoveConn()                      //уменьшает счётчик активных соединений
	GetConns() int64                  //получить количество активных соединений
	SetStatus(alive bool)             //установить статус сервера (доступен - не доступен)
	IsAlive() bool                    //получить статус сервера
	GetURL() string                   //получить URL сервера
	GetProxy() *httputil.ReverseProxy //получить reverse proxy сервера
}
