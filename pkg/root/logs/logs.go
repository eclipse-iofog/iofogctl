package logs

type logs struct {
	// TODO: (Serge) Add variables
}

func new() *logs {
	return &logs{}
}

func (logs *logs) validate(args []string) {
	println("Validate Logs")
	for idx := range args {
		println(args[idx])
	}
}

func (logs *logs) execute() {
	println("Execute Logs")
}