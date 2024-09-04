package interfaces

type BusContainer interface {
	RunCron() (err error)
}
