package docker_container

type Dockerinspect struct {
	Id             string
	Created        string
	Path           string
	Args           []string
	State          DockerState
	Image          string
	ResolvConfPath string
	HostnamePath   string
	HostsPath      string
	LogPath        string
	Name           string
	RestartCount   int
	Platform       string
	HostConfig     DockerHostConf
}

type DockerState struct {
	Status     string
	Running    bool
	Paused     bool
	Restarting bool
	Pid        int
	StartedAt  string
	FinishedAt string
}

type DockerHostConf struct {
	NetworkMode     string
	PortBindings    map[string]interface{}
	RestartPolicy   Restartpolicy
	PublishAllPorts bool
	CpuShares       int
	CpusetCpus      string
	PidsLimit       interface{}
	Mounts          []map[string]interface{}
}

type Restartpolicy struct {
	Name              string
	MaximumRetryCount int
}
