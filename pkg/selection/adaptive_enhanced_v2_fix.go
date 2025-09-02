package selection

// GetDetector returns the enhanced capability detector
func (eas *EnhancedAdaptiveSelector) GetDetector() *EnhancedCapabilityDetector {
	return eas.detector
}