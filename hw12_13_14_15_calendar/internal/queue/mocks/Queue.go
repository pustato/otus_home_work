// Code generated by mockery v2.10.2. DO NOT EDIT.

package mockqueue

import (
	queue "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/queue"
	mock "github.com/stretchr/testify/mock"
)

// Queue is an autogenerated mock type for the Queue type
type Queue struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Queue) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Connect provides a mock function with given fields:
func (_m *Queue) Connect() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateConsumer provides a mock function with given fields: exchange, _a1, key
func (_m *Queue) CreateConsumer(exchange string, _a1 string, key string) (queue.Consumer, error) {
	ret := _m.Called(exchange, _a1, key)

	var r0 queue.Consumer
	if rf, ok := ret.Get(0).(func(string, string, string) queue.Consumer); ok {
		r0 = rf(exchange, _a1, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Consumer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(exchange, _a1, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateProducer provides a mock function with given fields: exchange
func (_m *Queue) CreateProducer(exchange string) (queue.Producer, error) {
	ret := _m.Called(exchange)

	var r0 queue.Producer
	if rf, ok := ret.Get(0).(func(string) queue.Producer); ok {
		r0 = rf(exchange)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(queue.Producer)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(exchange)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
