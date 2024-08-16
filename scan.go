package ezdeploy

import (
	"os"
	"path/filepath"
)

const (
	SubdirHelm = "_helm"
)

type ScanResult struct {
	Charts     map[string]Chart
	Namespaces []string
}

func Scan(root string) (result ScanResult, err error) {
	// charts
	if result.Charts, err = scanCharts(filepath.Join(root, SubdirHelm)); err != nil {
		return
	}
	// namespaces
	if result.Namespaces, err = readDirNames(root); err != nil {
		return
	}
	return
}

func scanCharts(dir string) (charts map[string]Chart, err error) {
	charts = make(map[string]Chart)

	var names []string
	if names, err = readDirNames(dir); err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}

	for _, name := range names {
		chart := Chart{
			Name: name,
			Path: filepath.Join(dir, name),
		}
		if _, err = os.Stat(filepath.Join(chart.Path, "Chart.yaml")); err != nil {
			return
		}
		if _, err = os.Stat(filepath.Join(chart.Path, "values.yaml")); err != nil {
			return
		}
		if chart.Checksum, err = checksumDir(chart.Path); err != nil {
			return
		}
		charts[chart.Name] = chart
	}
	return
}
