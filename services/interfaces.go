// CreateGateway creates a new payment gateway instance
func (f *DefaultProviderFactory) CreateGateway(provider string, config map[string]interface{}) (PaymentGateway, error) {
	creator, exists := f.providers[provider]
	if !exists {
		return nil, &UnsupportedProviderError{Provider: provider}
	}
	return creator(config)
}