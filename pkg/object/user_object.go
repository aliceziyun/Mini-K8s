package object

type UserPod struct {
	Name   string
	Ready  string
	Status string
	Owner  string
	IP     string
	Ctime  string
}

type UserRS struct {
	Name    string
	Current int32
	Ctime   string
	Ready   int32
}

type UserNode struct {
	Name      string
	DynamicIP string
	Role      string
	Ctime     string
}
