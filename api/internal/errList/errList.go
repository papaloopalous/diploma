package errlist

var (
	ErrTemplate        = "Unable to execute the template"
	ErrDecrEmpty       = "Decryption data is empty"
	ErrDecrPaddingSize = "Invalid padding size"
	ErrDecrPaddindByte = "Invalid padding bytes"
	ErrDecrCipher      = "Ciphertext is not a multiple of the block size"
	ErrDecr            = "Failed to decrypt"
)
