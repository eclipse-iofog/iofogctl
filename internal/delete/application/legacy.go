package deleteapplication

func (exe *Executor) initLegacy() (err error) {
	application, err := exe.client.GetApplicationByName(exe.name)
	if err != nil {
		return
	}
	exe.application = application
	return
}

func (exe *Executor) deleteLegacy() (err error) {
	// Init remote resources
	if err = exe.initLegacy(); err != nil {
		return
	}

	// Delete application
	if err = exe.client.DeleteApplication(exe.application.Name); err != nil {
		return
	}
	return
}
