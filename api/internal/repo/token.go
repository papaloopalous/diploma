package repo

type TokenData struct {
	key []byte
}

type TokenRepo interface {
	GetData() (token []byte)
	SetData(token string)
}

var _ TokenRepo = &TokenData{}

func NewTokenRepo() *TokenData {
	return &TokenData{
		key: make([]byte, 0),
	}
}

func (p *TokenData) GetData() (token []byte) {
	res := p.key
	return res
}

func (p *TokenData) SetData(token string) {
	p.key = []byte(token)
}
