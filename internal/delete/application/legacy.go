package deleteapplication

func (exe *Executor) initLegacy() (err error) {
	flow, err := exe.client.GetFlowByName(exe.name)
	if err != nil {
		return
	}
	exe.flow = flow
	return
}

func (exe *Executor) deleteLegacy() (err error) {
	// Init remote resources
	if err = exe.initLegacy(); err != nil {
		return
	}

	// Delete flow
	if err = exe.client.DeleteFlow(exe.flow.ID); err != nil {
		return
	}
	return
}
