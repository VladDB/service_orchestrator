package process

// struct for process declaration and control
type Process struct {
	Id  int
	Pid int
}

func (u *Process) Start() bool {
	return true
}

func (u *Process) Stop() bool {
	return true
}

func (u *Process) Restart() bool {
	return true
}

func (u *Process) UpdateStatus() bool {
	return true
}
