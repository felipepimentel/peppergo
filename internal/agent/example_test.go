package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zaptest"

	"github.com/yourusername/peppergo/pkg/types"
)

// MockProvider is a mock implementation of types.Provider
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProvider) Generate(ctx context.Context, prompt string, opts ...types.GenerateOption) (*types.Response, error) {
	args := m.Called(ctx, prompt, opts)
	return args.Get(0).(*types.Response), args.Error(1)
}

func (m *MockProvider) Stream(ctx context.Context, prompt string) (<-chan types.Response, error) {
	args := m.Called(ctx, prompt)
	return args.Get(0).(<-chan types.Response), args.Error(1)
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) MaxTokens() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockProvider) SupportsStreaming() bool {
	args := m.Called()
	return args.Bool(0)
}

// MockCapability is a mock implementation of types.Capability
type MockCapability struct {
	mock.Mock
}

func (m *MockCapability) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCapability) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockCapability) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCapability) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0), args.Error(1)
}

func (m *MockCapability) Cleanup(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCapability) Requirements() *types.Requirements {
	args := m.Called()
	return args.Get(0).(*types.Requirements)
}

func (m *MockCapability) Version() string {
	args := m.Called()
	return args.String(0)
}

// MockTool is a mock implementation of types.Tool
type MockTool struct {
	mock.Mock
}

func (m *MockTool) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTool) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockTool) Initialize(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTool) Execute(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	mockArgs := m.Called(ctx, args)
	return mockArgs.Get(0), mockArgs.Error(1)
}

func (m *MockTool) Cleanup(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTool) Schema() *types.ToolSchema {
	args := m.Called()
	return args.Get(0).(*types.ToolSchema)
}

func (m *MockTool) Version() string {
	args := m.Called()
	return args.String(0)
}

func TestExampleAgent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	ctx := context.Background()

	t.Run("basic functionality", func(t *testing.T) {
		agent := NewExampleAgent(logger)
		assert.NotNil(t, agent)
		assert.Equal(t, "example-agent", agent.Name())
		assert.Equal(t, "1.0.0", agent.Version())
	})

	t.Run("execute with provider", func(t *testing.T) {
		agent := NewExampleAgent(logger)
		provider := new(MockProvider)
		expectedResponse := &types.Response{
			Content: "test response",
		}

		provider.On("Initialize", ctx).Return(nil)
		provider.On("Generate", ctx, "test task", mock.Anything).Return(expectedResponse, nil)
		provider.On("Name").Return("mock-provider")
		provider.On("MaxTokens").Return(1000)
		provider.On("SupportsStreaming").Return(true)

		err := agent.UseProvider(provider)
		assert.NoError(t, err)

		err = agent.Initialize(ctx)
		assert.NoError(t, err)

		response, err := agent.Execute(ctx, "test task")
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		provider.AssertExpectations(t)
	})

	t.Run("execute with capability", func(t *testing.T) {
		agent := NewExampleAgent(logger)
		capability := new(MockCapability)
		provider := new(MockProvider)
		expectedResponse := &types.Response{
			Content: "test response",
		}

		capability.On("Name").Return("test-capability")
		capability.On("Version").Return("1.0.0")
		capability.On("Initialize", ctx).Return(nil)
		capability.On("Execute", ctx, "test task").Return("capability result", nil)
		capability.On("Requirements").Return(types.NewRequirements())

		provider.On("Initialize", ctx).Return(nil)
		provider.On("Generate", ctx, "test task", mock.Anything).Return(expectedResponse, nil)

		err := agent.UseProvider(provider)
		assert.NoError(t, err)

		err = agent.AddCapability(capability)
		assert.NoError(t, err)

		err = agent.Initialize(ctx)
		assert.NoError(t, err)

		response, err := agent.Execute(ctx, "test task")
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		capability.AssertExpectations(t)
		provider.AssertExpectations(t)
	})

	t.Run("execute with tool", func(t *testing.T) {
		agent := NewExampleAgent(logger)
		tool := new(MockTool)
		provider := new(MockProvider)
		expectedResponse := &types.Response{
			Content: "test response",
		}

		tool.On("Name").Return("test-tool")
		tool.On("Version").Return("1.0.0")
		tool.On("Initialize", ctx).Return(nil)
		tool.On("Execute", ctx, mock.Anything).Return("tool result", nil)

		provider.On("Initialize", ctx).Return(nil)
		provider.On("Generate", ctx, "test task", mock.Anything).Return(expectedResponse, nil)

		err := agent.UseProvider(provider)
		assert.NoError(t, err)

		err = agent.AddTool(tool)
		assert.NoError(t, err)

		err = agent.Initialize(ctx)
		assert.NoError(t, err)

		response, err := agent.Execute(ctx, "test task")
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, response)

		tool.AssertExpectations(t)
		provider.AssertExpectations(t)
	})

	t.Run("custom setting", func(t *testing.T) {
		agent := NewExampleAgent(logger)
		assert.Equal(t, "default", agent.customSetting)

		agent.SetCustomSetting("new value")
		assert.Equal(t, "new value", agent.customSetting)
	})
} 