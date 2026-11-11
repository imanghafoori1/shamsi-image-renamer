package main

// runRename validates input, runs processDir, and refreshes the report.
func runRename(s *uiState) {
	if s.running {
		return
	}

	dir := s.dirEdit.Text()
	if dir == "" {
		s.showError(errorPathMessage())
		return
	}

	idx := s.formatCombo.CurrentIndex()
	if idx < 0 || idx >= len(dateFormats) {
		idx = 0
	}

	s.running = true
	s.statusLabel.SetText(startingMessage())

	prependModel := s.prependModelCheck.Checked()
	recursive := s.recursiveCheck.Checked()
	addTags := s.addTagsCheck.Checked()
	format := dateFormats[idx]

	go func() {
		report, err := processDir(dir, prependModel, format, recursive, addTags, s.setProgress)

		s.MainWindow.Synchronize(func() {
			s.running = false
			if err != nil {
				s.showError(errorPrefix(err.Error()))
				return
			}
			s.updateReport(report)
			s.statusLabel.SetText(doneTaggedMessage(len(report.Renamed), len(report.Skipped), len(report.Tagged)))
		})
	}()
}

// RemoveFunc is the operation runRemove invokes. It receives the directory
// and the recursive flag, and returns the report and any error.
type RemoveFunc func(dir string, recursive bool) (*Report, error)

// runRemove is the shared handler for the two cleanup buttons.
func runRemove(s *uiState, op RemoveFunc) {
	dir := s.dirEdit.Text()
	if dir == "" {
		s.showError(errorPathMessage())
		return
	}

	report, err := op(dir, s.recursiveCheck.Checked())
	if err != nil {
		s.statusLabel.SetText(errorPrefix(err.Error()))
		return
	}
	s.updateReport(report)
	s.statusLabel.SetText(doneMessage(len(report.Renamed), len(report.Skipped)))
}
