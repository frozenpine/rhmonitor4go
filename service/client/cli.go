package client

type Command string

const (
	Version Command = "version"
	Exit    Command = "exit"
	help    Command = "help"
)

type Cli struct {
	cmd  Command
	args []string
}

type CmdManager struct {
	api *RHRiskClient

	history []*Cli
}

func (m *CmdManager) Execute(cli *Cli) error {
	return nil
}
