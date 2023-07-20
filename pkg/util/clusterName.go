package util

import (
	"errors"
	"strings"
)

type ClusterName interface {
	IsNameValid(date, minHour string)
}

type Options struct {
	Name          string
	MaxNameLength int
}

func (opt *Options) IsNameValid() error {
	if len(opt.Name) > opt.MaxNameLength {
		return errors.New("cluster name is too long, please use a shorter name")
	}

	if strings.Contains(opt.Name, "live") || strings.Contains(opt.Name, "manager") {
		return errors.New("cluster name cannot contain the words 'live' or 'manager'")
	}
	return nil
}
