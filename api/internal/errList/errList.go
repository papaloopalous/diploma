package errlist

const (
	ErrTemplate        = "unable to execute the template"
	ErrDecrEmpty       = "decryption data is empty"
	ErrDecrPaddingSize = "invalid padding size"
	ErrDecrPaddindByte = "invalid padding bytes"
	ErrDecrCipher      = "ciphertext is not a multiple of the block size"
	ErrDecr            = "failed to decrypt"
	ErrInvalidToken    = "invalid token"
	ErrNoSession       = "no session was found"
	ErrInvalidLogin    = "username or password was incorrect"
	ErrUserNotFound    = "specified user was not found"
	ErrNameTaken       = "specified username is already taken"
	ErrSesNotFound     = "specified session was not found"
	ErrNoCookie        = "no cookie was found"
	ErrTokenParse      = "failed to parse a jwt token"
	ErrSesDelete       = "failed to delete a session"
	ErrNoPermission    = "you do not have permission"
	ErrNoTask          = "specified task was not found"
	ErrNoHeaders       = "missing headers"
	ErrReadingBody     = "failed to read the body"
)
