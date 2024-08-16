package ezdeploy

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/google/go-jsonnet"
	"gopkg.in/yaml.v3"
)

var (
	SuffixHelmValuesJSONNet = ".helm.jsonnet"

	SuffixesYAML       = FileSuffixes{".yaml", ".yml"}
	SuffixesJSON       = FileSuffixes{".json"}
	SuffixesJSONNet    = FileSuffixes{".jsonnet"}
	SuffixesHelmValues = FileSuffixes{".helm.yaml", ".helm.yml", SuffixHelmValuesJSONNet}
)

func sanitizeRawResources(raws *[]json.RawMessage) (err error) {
	for i := 0; i < len(*raws); i++ {
		// expand List
		var list List
		if err = json.Unmarshal((*raws)[i], &list); err != nil {
			return
		}
		if list.Valid() {
			*raws = append(append((*raws)[0:i], list.Items...), (*raws)[i+1:]...)
			// rewind if needed
			if len(list.Items) > 0 {
				i -= 1
			}
			continue
		}
		// re-marshal json
		var doc map[string]interface{}
		if err = json.Unmarshal((*raws)[i], &doc); err != nil {
			return
		}
		var buf []byte
		if buf, err = json.Marshal(doc); err != nil {
			return
		}
		(*raws)[i] = buf
	}

	return
}

func collectResourceFile(file string, namespace string) (raws []json.RawMessage, err error) {
	if SuffixesHelmValues.Match(file) {
		// ignore helm
	} else if SuffixesYAML.Match(file) {
		if err = collectYAMLFile(&raws, file); err != nil {
			return
		}
	} else if SuffixesJSON.Match(file) {
		if err = collectJSONFile(&raws, file); err != nil {
			return
		}
	} else if SuffixesJSONNet.Match(file) {
		if err = collectJSONNetFile(&raws, file, namespace); err != nil {
			return
		}
	} else {
		return
	}

	if err = sanitizeRawResources(&raws); err != nil {
		return
	}

	return
}

func collectYAMLFile(out *[]json.RawMessage, file string) (err error) {
	var raw []byte
	if raw, err = os.ReadFile(file); err != nil {
		return
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	for {
		var doc map[string]interface{}
		var buf []byte
		if err = dec.Decode(&doc); err == nil {
			if buf, err = json.Marshal(doc); err != nil {
				return
			}
			*out = append(*out, buf)
		} else {
			break
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func collectJSONContent(out *[]json.RawMessage, raw []byte) (err error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) < 2 {
		return
	}
	if raw[0] == '[' {
		var docs []json.RawMessage
		if err = json.Unmarshal(raw, &docs); err != nil {
			return
		}
		*out = append(*out, docs...)
	} else if raw[0] == '{' {
		*out = append(*out, raw)
	} else {
		sample := string(raw)
		if len(sample) > 10 {
			sample = sample[0:7] + "..."
		}
		err = errors.New("content is not a JSONObject or JSONArray: " + sample)
	}
	return
}

func collectJSONFile(out *[]json.RawMessage, file string) (err error) {
	var raw []byte
	if raw, err = os.ReadFile(file); err != nil {
		return
	}
	if err = collectJSONContent(out, raw); err != nil {
		return
	}
	return
}

func collectJSONNetFile(out *[]json.RawMessage, file string, namespace string) (err error) {
	vm := jsonnet.MakeVM()
	vm.ExtVar("NAMESPACE", namespace)
	var raw string
	if raw, err = vm.EvaluateFile(file); err != nil {
		return
	}
	if err = collectJSONContent(out, []byte(raw)); err != nil {
		return
	}
	return
}
