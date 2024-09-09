package damri

type util struct {
}

func NewUtil() *util {
	return &util{}
}

func (u *util) GetHMInMinutes(hours int, minutes int) int {
	return hours*60 + minutes
}
