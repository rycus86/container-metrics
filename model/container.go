package model

type Container struct {
	Id     string
	Name   string
	Image  string
	Labels map[string]string

	ComputedLabels map[string]string
}
