func processDir(dir string) (*Report, error) {
	report := &Report{}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return report, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".jpg" && ext != ".jpeg" {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		f, err := os.Open(fullPath)
		if err != nil {
			report.addSkipped(entry.Name(), "cannot open")
			continue
		}

		x, err := exif.Decode(f)
		f.Close()
		if err != nil || x == nil {
			report.addSkipped(entry.Name(), "no EXIF data")
			continue
		}

		tag, err := x.Get(exif.DateTimeOriginal)
		if err != nil {
			report.addSkipped(entry.Name(), "no DateTimeOriginal")
			continue
		}

		dateStr, err := tag.StringVal()
		if err != nil {
			report.addSkipped(entry.Name(), "unreadable date tag")
			continue
		}

		tm, err := time.Parse("2006:01:02 15:04:05", dateStr)
		if err != nil {
			report.addSkipped(entry.Name(), "bad date format")
			continue
		}

		jy, jm, _, _ := jalaali.ToJalaali(tm.Year(), tm.Month(), tm.Day())
		monthName := getPersianMonth(jm)

		suffix := fmt.Sprintf("-%s-%d", monthName, jy)
		fileExt := filepath.Ext(entry.Name())
		nameWithoutExt := strings.TrimSuffix(entry.Name(), fileExt)

		// Already carries exactly this suffix -> leave it alone.
		if strings.HasSuffix(nameWithoutExt, suffix) {
			report.addSkipped(entry.Name(), "already renamed")
			continue
		}

		newName := nameWithoutExt + suffix + fileExt
		newPath := filepath.Join(dir, newName)

		if _, err := os.Stat(newPath); err == nil {
			report.addSkipped(entry.Name(), "target already exists")
			continue
		}

		if err := os.Rename(fullPath, newPath); err != nil {
			report.addSkipped(entry.Name(), "rename failed")
			continue
		}
		report.addRenamed(entry.Name(), newName)
	}
	return report, nil
}