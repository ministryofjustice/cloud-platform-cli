// Code generated by mockery v2.46.1. DO NOT EDIT.

package mocks

import (
	context "context"
	io "io"

	mock "github.com/stretchr/testify/mock"

	tfexec "github.com/hashicorp/terraform-exec/tfexec"

	tfjson "github.com/hashicorp/terraform-json"
)

// TerraformExec is an autogenerated mock type for the terraformExec type
type TerraformExec struct {
	mock.Mock
}

// Apply provides a mock function with given fields: ctx, opts
func (_m *TerraformExec) Apply(ctx context.Context, opts ...tfexec.ApplyOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Apply")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.ApplyOption) error); ok {
		r0 = rf(ctx, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Destroy provides a mock function with given fields: ctx, opts
func (_m *TerraformExec) Destroy(ctx context.Context, opts ...tfexec.DestroyOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Destroy")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.DestroyOption) error); ok {
		r0 = rf(ctx, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Init provides a mock function with given fields: ctx, opts
func (_m *TerraformExec) Init(ctx context.Context, opts ...tfexec.InitOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Init")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.InitOption) error); ok {
		r0 = rf(ctx, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Output provides a mock function with given fields: ctx, opts
func (_m *TerraformExec) Output(ctx context.Context, opts ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Output")
	}

	var r0 map[string]tfexec.OutputMeta
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.OutputOption) (map[string]tfexec.OutputMeta, error)); ok {
		return rf(ctx, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.OutputOption) map[string]tfexec.OutputMeta); ok {
		r0 = rf(ctx, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]tfexec.OutputMeta)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...tfexec.OutputOption) error); ok {
		r1 = rf(ctx, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Plan provides a mock function with given fields: ctx, opts
func (_m *TerraformExec) Plan(ctx context.Context, opts ...tfexec.PlanOption) (bool, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Plan")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.PlanOption) (bool, error)); ok {
		return rf(ctx, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.PlanOption) bool); ok {
		r0 = rf(ctx, opts...)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...tfexec.PlanOption) error); ok {
		r1 = rf(ctx, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetStderr provides a mock function with given fields: w
func (_m *TerraformExec) SetStderr(w io.Writer) {
	_m.Called(w)
}

// SetStdout provides a mock function with given fields: w
func (_m *TerraformExec) SetStdout(w io.Writer) {
	_m.Called(w)
}

// Show provides a mock function with given fields: ctx, opts
func (_m *TerraformExec) Show(ctx context.Context, opts ...tfexec.ShowOption) (*tfjson.State, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for Show")
	}

	var r0 *tfjson.State
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.ShowOption) (*tfjson.State, error)); ok {
		return rf(ctx, opts...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...tfexec.ShowOption) *tfjson.State); ok {
		r0 = rf(ctx, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tfjson.State)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...tfexec.ShowOption) error); ok {
		r1 = rf(ctx, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WorkspaceDelete provides a mock function with given fields: ctx, workspace, opts
func (_m *TerraformExec) WorkspaceDelete(ctx context.Context, workspace string, opts ...tfexec.WorkspaceDeleteCmdOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, workspace)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for WorkspaceDelete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...tfexec.WorkspaceDeleteCmdOption) error); ok {
		r0 = rf(ctx, workspace, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WorkspaceNew provides a mock function with given fields: ctx, workspace, opts
func (_m *TerraformExec) WorkspaceNew(ctx context.Context, workspace string, opts ...tfexec.WorkspaceNewCmdOption) error {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, workspace)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	if len(ret) == 0 {
		panic("no return value specified for WorkspaceNew")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, ...tfexec.WorkspaceNewCmdOption) error); ok {
		r0 = rf(ctx, workspace, opts...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WorkspaceSelect provides a mock function with given fields: ctx, workspace
func (_m *TerraformExec) WorkspaceSelect(ctx context.Context, workspace string) error {
	ret := _m.Called(ctx, workspace)

	if len(ret) == 0 {
		panic("no return value specified for WorkspaceSelect")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, workspace)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewTerraformExec creates a new instance of TerraformExec. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTerraformExec(t interface {
	mock.TestingT
	Cleanup(func())
}) *TerraformExec {
	mock := &TerraformExec{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
