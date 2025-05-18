package encryption

import (
	"api/internal/messages"
	"api/internal/response"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log"
	"math/big"
	"net/http"

	"github.com/spf13/viper"
)

// Криптографические параметры для обмена ключами по схеме Диффи-Хеллмана
var (
	prime     *big.Int
	generator *big.Int

	serverPriv, _ = new(big.Int).SetString("1234567890ABCDEF1234567890ABCDEF12345678", 16)
	serverPub     = new(big.Int).Exp(generator, serverPriv, prime)
)

func init() {
	prime, _ = new(big.Int).SetString(viper.GetString("crypto.prime"), 16)
	generator = big.NewInt(viper.GetInt64("crypto.generator"))
}

// GetCryptoParams отправляет параметры для установки защищенного соединения
func GetCryptoParams(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		messages.CryptoParamPrime:     prime.String(),
		messages.CryptoParamGenerator: generator.String(),
	}

	log.Printf(messages.LogStatusParamsSent, prime.String(), generator.String())
	response.WriteAPIResponse(w, http.StatusOK, true, "", params)
}

// GetServerPublicKey возвращает публичный ключ сервера
func GetServerPublicKey() string {
	return serverPub.String()
}

// DeriveSharedKeyHex вычисляет общий секретный ключ по схеме Диффи-Хеллмана
func DeriveSharedKeyHex(clientPublic string) (string, error) {
	cliPub, ok := new(big.Int).SetString(clientPublic, 10)
	if !ok {
		log.Printf(messages.LogErrInvalidPublicKey, clientPublic)
		return "", errors.New(messages.ClientErrInvalidPublicKey)
	}

	secret := new(big.Int).Exp(cliPub, serverPriv, prime)
	decStr := secret.String()
	hash := sha256.Sum256([]byte(decStr))

	log.Printf(messages.LogStatusKeyDerived, clientPublic)
	return hex.EncodeToString(hash[:]), nil
}
