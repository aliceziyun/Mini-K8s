package object

type UserPod struct {
	Name   string
	Ready  string
	Status string
	IP     string
}

type UserRS struct {
	Name    string
	Current int32
	Ready   int32
}
