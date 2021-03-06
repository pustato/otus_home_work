// Code generated by mockery v2.10.2. DO NOT EDIT.

package mockqueue

import (
	queue "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/queue"
	mock "github.com/stretchr/testify/mock"
)

// Producer is an autogenerated mock type for the Producer type
type Producer struct {
	mock.Mock
}

// Publish provides a mock function with given fields: m
func (_m *Producer) Publish(m *queue.Message) error {
	ret := _m.Called(m)

	var r0 error
	if rf, ok := ret.Get(0).(func(*queue.Message) error); ok {
		r0 = rf(m)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
