package ezops

import (
	"encoding/json"
	"github.com/karrick/godirwalk"
	"path/filepath"
	"strings"
)

type LoadResult struct {
	Releases     []Release
	Resources    []Resource
	ResourcesExt []Resource
}

type LoadOptions struct {
	Charts map[string]Chart
}

func Load(root string, namespace string, opts LoadOptions) (result LoadResult, err error) {
	if err = godirwalk.Walk(filepath.Join(root, namespace), &godirwalk.Options{
		FollowSymbolicLinks: true,
		Callback: func(file string, entry *godirwalk.Dirent) (err error) {
			if entry.IsDir() ||
				strings.HasPrefix(entry.Name(), ".") ||
				strings.HasPrefix(entry.Name(), "_") {
				return
			}

			var raws []json.RawMessage
			if raws, err = collectResourceFile(file, namespace); err != nil {
				return
			}

			if len(raws) == 0 {
				return
			}

			for _, raw := range raws {
				res := Resource{
					Namespace: namespace,
					Raw:       raw,
					Path:      file,
					Checksum:  checksumBytes(raw),
				}

				if err = json.Unmarshal(raw, &res.Object); err != nil {
					return
				}

				res.ID = CreateResourceID(namespace, res.Object)

				if res.Object.Metadata.Namespace == "" {
					result.Resources = append(result.Resources, res)
				} else {
					result.ResourcesExt = append(result.ResourcesExt, res)
				}
			}
			return
		},
	}); err != nil {
		return
	}

	if result.Releases, err = collectReleases(root, namespace, opts.Charts); err != nil {
		return
	}

	return
}
