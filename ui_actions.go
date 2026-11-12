package main

// workFunc is the long-running operation both runRename and runRemove
// hand to runInBackground. It's called on the worker goroutine and must
// not touch the GUI.
type workFunc func() (*Report, error)

// runInBackground handles the shared lifecycle: guard against re-entry,
// flip the running flag, start the worker, and marshal the result back
// onto the GUI thread.
func runInBackground(s *uiState, message func(*Report) string, work workFunc) {
	if s.running {
		return
	}
	s.running = true
	s.statusLabel.SetText(startingMessage())

	go func() {
		report, err := work()
		s.MainWindow.Synchronize(func() {
			s.running = false
			if err != nil {
				s.showError(errorPrefix(err.Error()))
				return
			}
			s.updateReport(report)
			s.statusLabel.SetText(message(report))
		})
	}()
}

// runRename validates input and kicks off processDir.
func runRename(s *uiState, message func(*Report) string) {
	options := readFormState(s)
	if options.folders.rootDir == "" {
		s.showError(errorPathMessage())
		return
	}

	runInBackground(s, message, func() (*Report, error) {
		return processDir(options, s.setProgress)
	})
}

// RemoveFunc is the operation runRemove invokes. It receives the directory
// and the recursive flag, and returns the report and any error.
type RemoveFunc func(dir folders) (*Report, error)

func runRemove(s *uiState, op RemoveFunc, message func(*Report) string) {
	folders := readFormState(s).folders
	if folders.rootDir == "" {
		s.showError(errorPathMessage())
		return
	}

	runInBackground(s, message, func() (*Report, error) {
		return op(folders)
	})
}
