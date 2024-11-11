package vo

type LoginTokenList struct {
	Token           string
	TokenExpireSecs int
	ErrorCount      int
	TotalAttempts   int
}
