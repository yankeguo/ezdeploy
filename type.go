package ezdeploy

import "encoding/json"

// Helm

type Chart struct {
	Name     string
	Path     string
	Checksum string
}

type Release struct {
	ID         string
	Name       string
	Chart      Chart
	ValuesFile string
	Checksum   string
}

func CreateReleaseID(namespace string, name string) string {
	return namespace + "::" + "Helm" + "::" + name
}

// List

type List struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Items      []json.RawMessage `json:"items"`
}

func NewList(items []json.RawMessage) *List {
	return &List{
		APIVersion: "v1",
		Kind:       "List",
		Items:      items,
	}
}

func (l List) Valid() bool {
	return l.APIVersion == "v1" && l.Kind == "List"
}

// Resource

type Resource struct {
	ID        string
	Namespace string
	Object    Object
	Raw       json.RawMessage
	Checksum  string
	Path      string
}

type ObjectMeta struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type Object struct {
	Metadata   ObjectMeta `json:"metadata"`
	Kind       string     `json:"kind"`
	APIVersion string     `json:"apiVersion"`
}

func CreateResourceID(namespace string, object Object) string {
	if object.Metadata.Namespace != "" {
		namespace = object.Metadata.Namespace
	}
	return namespace + "::" + object.APIVersion + "/" + object.Kind + "/" + object.Metadata.Name
}
