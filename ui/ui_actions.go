package ui

// WorkFunc is the long-running operation both runRename and runRemove
// hand to runInBackground. It's called on the worker goroutine and must
// not touch the GUI.
type WorkFunc func() (*Report, error)
type ProgressFunc func(done, total int, currentName string)

// runInBackground handles the shared lifecycle: guard against re-entry,
// flip the running flag, start the worker, and marshal the result back
// onto the GUI thread.
func (h *actionHandler) runInBackground(work WorkFunc) {
	s := h.uiState
	if s.running {
		return
	}
	s.running = true
	s.StatusLabel.SetText(startingMessage())

	go func() {
		report, err := work()
		s.MainWindow.Synchronize(func() {
			s.running = false
			if err != nil {
				s.showError(errorPrefix(err.Error()))
				return
			}
			s.updateReport(report)
			s.StatusLabel.SetText(h.message(report))
		})
	}()
}

// runRename validates input and kicks off processDir.
func runRename(h actionHandler, cb Renamer) {
	opts := readFormState(h.uiState)
	if opts.Folders.RootDir == "" {
		h.uiState.showError(errorPathMessage())
		return
	}

	h.runInBackground(cb(opts, h.uiState.setProgress))
}

// RemoveFunc is the operation runRemove invokes. It receives the directory
// and the recursive flag, and returns the report and any error.
type RemoveFunc func(dir Folders) (*Report, error)

type RenameFunc func() (*Report, error)

func (h *actionHandler) runRemove(op RemoveFunc) {
	s := h.uiState
	folders := readFormState(s).Folders
	if folders.RootDir == "" {
		s.showError(errorPathMessage())
		return
	}

	h.runInBackground(func() (*Report, error) {
		return op(folders)
	})
}
