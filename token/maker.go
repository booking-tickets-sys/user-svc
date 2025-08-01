package token

type TokenMaker interface {
	CreateToken(username string, duration int64) (string, error)
	CreateRefreshToken(username string, duration int64) (string, error)
	VerifyToken(token string) (*Payload, error)
	VerifyRefreshToken(token string) (*Payload, error)
}
