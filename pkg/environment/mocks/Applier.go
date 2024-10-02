// Code generated by mockery v2.46.1. DO NOT EDIT.

package mocks

import (
	tfjson "github.com/hashicorp/terraform-json"
	mock "github.com/stretchr/testify/mock"
)

// Applier is an autogenerated mock type for the Applier type
type Applier struct {
	mock.Mock
}

// Initialize provides a mock function with given fields:
func (_m *Applier) Initialize() {
	_m.Called()
}

// KubectlApply provides a mock function with given fields: namespace, directory, dryRun
func (_m *Applier) KubectlApply(namespace string, directory string, dryRun bool) (string, error) {
	ret := _m.Called(namespace, directory, dryRun)

	if len(ret) == 0 {
		panic("no return value specified for KubectlApply")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, bool) (string, error)); ok {
		return rf(namespace, directory, dryRun)
	}
	if rf, ok := ret.Get(0).(func(string, string, bool) string); ok {
		r0 = rf(namespace, directory, dryRun)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string, bool) error); ok {
		r1 = rf(namespace, directory, dryRun)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// KubectlDelete provides a mock function with given fields: namespace, directory, dryRun
func (_m *Applier) KubectlDelete(namespace string, directory string, dryRun bool) (string, error) {
	ret := _m.Called(namespace, directory, dryRun)

	if len(ret) == 0 {
		panic("no return value specified for KubectlDelete")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, bool) (string, error)); ok {
		return rf(namespace, directory, dryRun)
	}
	if rf, ok := ret.Get(0).(func(string, string, bool) string); ok {
		r0 = rf(namespace, directory, dryRun)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string, bool) error); ok {
		r1 = rf(namespace, directory, dryRun)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TerraformDestroy provides a mock function with given fields: directory
func (_m *Applier) TerraformDestroy(directory string) error {
	ret := _m.Called(directory)

	if len(ret) == 0 {
		panic("no return value specified for TerraformDestroy")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(directory)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TerraformInitAndApply provides a mock function with given fields: namespace, directory
func (_m *Applier) TerraformInitAndApply(namespace string, directory string) (string, error) {
	ret := _m.Called(namespace, directory)

	if len(ret) == 0 {
		panic("no return value specified for TerraformInitAndApply")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (string, error)); ok {
		return rf(namespace, directory)
	}
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(namespace, directory)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(namespace, directory)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TerraformInitAndDestroy provides a mock function with given fields: namespace, directory
func (_m *Applier) TerraformInitAndDestroy(namespace string, directory string) (string, error) {
	ret := _m.Called(namespace, directory)

	if len(ret) == 0 {
		panic("no return value specified for TerraformInitAndDestroy")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (string, error)); ok {
		return rf(namespace, directory)
	}
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(namespace, directory)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(namespace, directory)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TerraformInitAndPlan provides a mock function with given fields: namespace, directory
func (_m *Applier) TerraformInitAndPlan(namespace string, directory string) (*tfjson.Plan, string, error) {
	ret := _m.Called(namespace, directory)

	if len(ret) == 0 {
		panic("no return value specified for TerraformInitAndPlan")
	}

	var r0 *tfjson.Plan
	var r1 string
	var r2 error
	if rf, ok := ret.Get(0).(func(string, string) (*tfjson.Plan, string, error)); ok {
		return rf(namespace, directory)
	}
	if rf, ok := ret.Get(0).(func(string, string) *tfjson.Plan); ok {
		r0 = rf(namespace, directory)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tfjson.Plan)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) string); ok {
		r1 = rf(namespace, directory)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(string, string) error); ok {
		r2 = rf(namespace, directory)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// NewApplier creates a new instance of Applier. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewApplier(t interface {
	mock.TestingT
	Cleanup(func())
}) *Applier {
	mock := &Applier{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
