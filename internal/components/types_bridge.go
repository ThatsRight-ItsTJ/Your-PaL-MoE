package components

import (
	"github.com/ThatsRight-ItsTJ/Your-PaL-MoE/internal/enhanced"
)

// TypesBridge provides compatibility between components and enhanced packages
type TypesBridge struct{}

// NewTypesBridge creates a new types bridge
func NewTypesBridge() *TypesBridge {
	return &TypesBridge{}
}

// ConvertComplexity converts complexity levels between packages
func (tb *TypesBridge) ConvertComplexity(level enhanced.ComplexityLevel) enhanced.ComplexityLevel {
	return level
}

// ConvertTaskComplexity converts task complexity between packages
func (tb *TypesBridge) ConvertTaskComplexity(complexity *enhanced.TaskComplexity) *enhanced.TaskComplexity {
	return complexity
}

// ConvertProvider converts provider between packages
func (tb *TypesBridge) ConvertProvider(provider *enhanced.Provider) *enhanced.Provider {
	return provider
}

// ConvertProviderAssignment converts provider assignment between packages
func (tb *TypesBridge) ConvertProviderAssignment(assignment *enhanced.ProviderAssignment) *enhanced.ProviderAssignment {
	return assignment
}

// ConvertRequestInput converts request input between packages
func (tb *TypesBridge) ConvertRequestInput(input *enhanced.RequestInput) *enhanced.RequestInput {
	return input
}

// ConvertProcessResponse converts process response between packages
func (tb *TypesBridge) ConvertProcessResponse(response *enhanced.ProcessResponse) *enhanced.ProcessResponse {
	return response
}