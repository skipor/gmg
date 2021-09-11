package example

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mocks_simple_mock_usage "github.com/skipor/gmg/examples/1_simple_mock_usage/mocks"
)

func TestDo(t *testing.T) {
	// When Do called
	// Then Bar called And its argument contains "something"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	foo := mocks_simple_mock_usage.NewMockFoo(ctrl)
	foo.EXPECT().Bar(gomock.Any()).DoAndReturn(func(s string) error {
		assert.Contains(t, s, "something")
		return nil
	})
	err := Do(foo)
	require.NoError(t, err)
}
