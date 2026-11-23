package ui

import . "renamer/report"
import . "renamer/dirWalker"

// WorkFunc is the long-running operation both runRename and runRemove
// hand to runInBackground. It's called on the worker goroutine and must
// not touch the GUI.
type WorkFunc func() (*Report, error)

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

// validates input and kicks off processDir.
func (h actionHandler) runRename(cb renamerFunc) {
	opts := readFormState(h.uiState)
	if opts.Folders.RootDir == "" {
		h.uiState.showError(errorPathMessage())
		return
	}

	h.runInBackground(cb(opts, h.uiState.setProgress))
}

// removeFunc is the operation runRemove invokes. It receives the directory
// and the recursive flag, and returns the report and any error.
type removeFunc func(dir Folders) (*Report, error)

//type RenameFunc func() (*Report, error)

func (h *actionHandler) runRemove(op removeFunc) {
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
