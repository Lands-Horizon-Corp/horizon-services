package secret

type Secret[TClaim any] interface {
	GetToken() (*TClaim, error)
	CleanToken()
	VerifyToken()
	GenerateToken()
}
