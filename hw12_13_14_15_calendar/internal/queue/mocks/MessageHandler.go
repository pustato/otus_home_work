// Code generated by mockery v2.10.2. DO NOT EDIT.

package mockqueue

import (
	queue "github.com/pustato/otus_home_work/hw12_13_14_15_calendar/internal/queue"
	mock "github.com/stretchr/testify/mock"
)

// MessageHandler is an autogenerated mock type for the MessageHandler type
type MessageHandler struct {
	mock.Mock
}

// Execute provides a mock function with given fields: m
func (_m *MessageHandler) Execute(m *queue.Message) {
	_m.Called(m)
}
