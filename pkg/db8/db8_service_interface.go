package db8

type Db8Service8Interface interface {
	GetUniqueServices() ([]string, error)
	CountUniqueServices() (int, error)
}
