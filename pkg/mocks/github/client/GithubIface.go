// Code generated by mockery v2.14.0. DO NOT EDIT.

package mocks

import (
	github "github.com/google/go-github/github"
	client "github.com/ministryofjustice/cloud-platform-cli/pkg/github/client"

	mock "github.com/stretchr/testify/mock"

	util "github.com/ministryofjustice/cloud-platform-cli/pkg/util"
)

// GithubIface is an autogenerated mock type for the GithubIface type
type GithubIface struct {
	mock.Mock
}

// GetChangedFiles provides a mock function with given fields: _a0
func (_m *GithubIface) GetChangedFiles(_a0 int) ([]*github.CommitFile, error) {
	ret := _m.Called(_a0)

	var r0 []*github.CommitFile
	if rf, ok := ret.Get(0).(func(int) []*github.CommitFile); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.CommitFile)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListMergedPRs provides a mock function with given fields: date, count
func (_m *GithubIface) ListMergedPRs(date util.Date, count int) ([]client.Nodes, error) {
	ret := _m.Called(date, count)

	var r0 []client.Nodes
	if rf, ok := ret.Get(0).(func(util.Date, int) []client.Nodes); ok {
		r0 = rf(date, count)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]client.Nodes)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(util.Date, int) error); ok {
		r1 = rf(date, count)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewGithubIface interface {
	mock.TestingT
	Cleanup(func())
}

// NewGithubIface creates a new instance of GithubIface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewGithubIface(t mockConstructorTestingTNewGithubIface) *GithubIface {
	mock := &GithubIface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
