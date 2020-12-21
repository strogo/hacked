package archives

type viewModel struct {
	windowOpen   bool
	restoreFocus bool

	selectedLevel int

	emailIndex    int
	logIndex      int
	fragmentIndex int
}

func freshViewModel() viewModel {
	return viewModel{}
}
