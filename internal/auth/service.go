package auth

// Service provides password hashing, JWT issue/verify, and API token generation.
type Service struct {
	jwtSecret string
}

// New creates a Service. jwtSecret should be a long random string from config.
func New(jwtSecret string) *Service {
	return &Service{jwtSecret: jwtSecret}
}

func (s *Service) HashPassword(pw string) (string, error)    { return hashPassword(pw) }
func (s *Service) CheckPassword(pw, hash string) bool        { return checkPassword(pw, hash) }
func (s *Service) GenerateAPIToken() (string, error)         { return generateAPIToken() }
