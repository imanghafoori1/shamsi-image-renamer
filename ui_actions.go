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

// runRemove is the shared handler for the two cleanup buttons.
func runRemove(s *uiState, stripPrefixes, stripSuffixes bool) {
	dir := s.dirEdit.Text()
	if dir == "" {
		s.showError(errorPathMessage())
		return
	}

	var report *Report
	var err error
	switch {
	case stripPrefixes:
		report, err = removeModelPrefixes(dir, s.recursiveCheck.Checked())
	case stripSuffixes:
		report, err = removeDateSuffixes(dir, s.recursiveCheck.Checked())
	}
	if err != nil {
		s.statusLabel.SetText(errorPrefix(err.Error()))
		return
	}
	s.updateReport(report)
	s.statusLabel.SetText(doneMessage(len(report.Renamed), len(report.Skipped)))
}
