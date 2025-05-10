package db8

type Db8Hostnameinfo8Interface interface {
	GetUniqueSoftware() ([]string, error)
	CountUniqueSoftware() (int, error)
}
