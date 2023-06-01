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

type UserService struct {
	Name      string
	NameSpace string
	Selector  map[string]string
	IPFamily  string
	IP        string
	Port      []string
	EndPoint  []string
}

type UserJob struct {
	Name   string
	Status string
	Ctime  string
}

type UserFunc struct {
	Name     string
	Type     string
	FuncName string
	Path     string
}
