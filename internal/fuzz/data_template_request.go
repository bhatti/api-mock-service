package fuzz

// DataTemplateRequest for generating random data pattern
type DataTemplateRequest struct {
	IncludeType   bool `yaml:"include_type" json:"include_type"`
	MinMultiplier int  `yaml:"min_multiplier" json:"min_multiplier"`
	MaxMultiplier int  `yaml:"max_multiplier" json:"max_multiplier"`
}

// NewDataTemplateRequest constructor
func NewDataTemplateRequest(includeType bool, minMultiplier int, maxMultiplier int) DataTemplateRequest {
	return DataTemplateRequest{
		IncludeType:   includeType,
		MinMultiplier: minMultiplier,
		MaxMultiplier: maxMultiplier,
	}
}

// WithInclude instantiates copy with a new include type
func (dtr DataTemplateRequest) WithInclude(includeType bool) DataTemplateRequest {
	return DataTemplateRequest{
		IncludeType:   includeType,
		MinMultiplier: dtr.MinMultiplier,
		MaxMultiplier: dtr.MaxMultiplier,
	}
}

// WithMaxMultiplier instantiates copy with a new max multiplier
func (dtr DataTemplateRequest) WithMaxMultiplier(multiplier int) DataTemplateRequest {
	return DataTemplateRequest{
		IncludeType:   dtr.IncludeType,
		MinMultiplier: dtr.MinMultiplier,
		MaxMultiplier: multiplier,
	}
}
