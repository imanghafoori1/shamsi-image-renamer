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
	dir := s.dirEdit.Text()
	if dir == "" {
		s.showError(errorPathMessage())
		return
	}

	idx := s.formatCombo.CurrentIndex()
	if idx < 0 || idx >= len(dateFormats) {
		idx = 0
	}

	prependModel := s.prependModelCheck.Checked()
	recursive := s.recursiveCheck.Checked()
	addTags := s.addTagsCheck.Checked()
	format := dateFormats[idx]

	runInBackground(s, message, func() (*Report, error) {
		return processDir(dir, prependModel, format, recursive, addTags, s.setProgress)
	})
}

// RemoveFunc is the operation runRemove invokes. It receives the directory
// and the recursive flag, and returns the report and any error.
type RemoveFunc func(dir string, recursive bool) (*Report, error)

func runRemove(s *uiState, op RemoveFunc, message func(*Report) string) {
	dir := s.dirEdit.Text()
	if dir == "" {
		s.showError(errorPathMessage())
		return
	}
	recursive := s.recursiveCheck.Checked() // read on GUI thread

	runInBackground(s, message, func() (*Report, error) {
		return op(dir, recursive)
	})
}
