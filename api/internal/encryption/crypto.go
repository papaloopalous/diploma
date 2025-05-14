package encryption

import (
	"api/internal/response"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"net/http"
)

var (
	prime, _  = new(big.Int).SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD1", 16)
	generator = big.NewInt(2)

	serverPriv, _ = new(big.Int).SetString("1234567890ABCDEF1234567890ABCDEF12345678", 16)
	serverPub     = new(big.Int).Exp(generator, serverPriv, prime)
)

func GetCryptoParams(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{
		"prime":     "FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD1",
		"generator": "2",
	}

	response.WriteAPIResponse(w, http.StatusOK, true, "", params)
}

func GetServerPublicKey() string {
	return serverPub.String()
}

func DeriveSharedKeyHex(clientPublic string) (string, error) {
	cliPub, ok := new(big.Int).SetString(clientPublic, 10)
	if !ok {
		return "", errors.New("invalid client public key")
	}

	secret := new(big.Int).Exp(cliPub, serverPriv, prime)
	decStr := secret.String()
	hash := sha256.Sum256([]byte(decStr))
	return hex.EncodeToString(hash[:]), nil
}
