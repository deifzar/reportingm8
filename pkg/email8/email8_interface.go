package email8

type Email8Interface interface {
	SendEmail(recipients []string, message []byte) error
}
