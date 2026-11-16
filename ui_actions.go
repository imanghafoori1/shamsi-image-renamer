package main

// workFunc is the long-running operation both runRename and runRemove
// hand to runInBackground. It's called on the worker goroutine and must
// not touch the GUI.
type workFunc func() (*Report, error)

type actionHandler struct {
	uiState *uiState
	message func(*Report) string
}

// runInBackground handles the shared lifecycle: guard against re-entry,
// flip the running flag, start the worker, and marshal the result back
// onto the GUI thread.
func (h *actionHandler) runInBackground(work workFunc) {
	s := h.uiState
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
			s.statusLabel.SetText(h.message(report))
		})
	}()
}

// runRename validates input and kicks off processDir.
func (h *actionHandler) runRename() {
	options := readFormState(h.uiState)
	if options.folders.rootDir == "" {
		h.uiState.showError(errorPathMessage())
		return
	}

	h.runInBackground(func() (*Report, error) {
		return processDir(options, h.uiState.setProgress)
	})
}

// RemoveFunc is the operation runRemove invokes. It receives the directory
// and the recursive flag, and returns the report and any error.
type RemoveFunc func(dir folders) (*Report, error)

func (h *actionHandler) runRemove(op RemoveFunc) {
	s := h.uiState
	folders := readFormState(s).folders
	if folders.rootDir == "" {
		s.showError(errorPathMessage())
		return
	}

	h.runInBackground(func() (*Report, error) {
		return op(folders)
	})
}
