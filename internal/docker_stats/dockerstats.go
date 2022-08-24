package dockerstats

const (
	defaultDockerPath        string = "docker"
	defaultDockerCommand     string = "stats"
	defaultDockerAll         string = "-a"
	defaultDockerNoStreamArg string = "--no-stream"
	defaultDockerFormatArg   string = "--format"
	defaultDockerFormat      string = `{"container":"{{.Container}}","memory":{"raw":"{{.MemUsage}}","percent":"{{.MemPerc}}"},"cpu":"{{.CPUPerc}}","io":{"network":"{{.NetIO}}","block":"{{.BlockIO}}"},"pids":{{.PIDs}},"name":"{{.Name}}"}`
)

// DefaultCommunicator is the default way of retrieving stats from Docker.
//
// When calling `Current()`, the `DefaultCommunicator` is used, and when
// retriving a `Monitor` using `NewMonitor()`, it is initialized with the
// `DefaultCommunicator`.
var DefaultCommunicator Communicator = CliCommunicator{
	DockerPath: defaultDockerPath,
	Command:    []string{defaultDockerCommand, defaultDockerAll, defaultDockerNoStreamArg, defaultDockerFormatArg, defaultDockerFormat},
}

// Current returns the current `Stats` of each running Docker container.
//
// Current will always return a `[]Stats` slice equal in length to the number of
// running Docker containers, or an `error`. No error is returned if there are no
// running Docker containers, simply an empty slice.
func Current() ([]Stats, error) {
	return DefaultCommunicator.Stats()
}
