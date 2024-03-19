package ezops

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func collectReleases(root string, namespace string, charts map[string]Chart) (releases []Release, err error) {
	dir := filepath.Join(root, namespace)

	var entries []fs.DirEntry
	if entries, err = os.ReadDir(dir); err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !SuffixesHelmValues.Match(entry.Name()) {
			continue
		}
		splits := strings.SplitN(entry.Name(), ".", 3)
		if len(splits) != 3 {
			continue
		}
		name, chartName := splits[0], splits[1]
		chart, ok := charts[chartName]
		if !ok {
			err = errors.New("missing chart named '" + chartName + "'")
			return
		}
		release := Release{
			ID:         CreateReleaseID(namespace, name),
			Name:       name,
			Chart:      chart,
			ValuesFile: filepath.Join(dir, entry.Name()),
		}

		var checksum string
		if checksum, err = checksumFile(release.ValuesFile); err != nil {
			return
		}
		release.Checksum = checksumBytes([]byte(chart.Checksum + checksum))

		releases = append(releases, release)
	}
	return
}
