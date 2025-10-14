package kairyu

type aggregator interface {
	aggregate(responses [][]byte, mode string) []byte
}

type defaultAggregator struct{}

func (a *defaultAggregator) aggregate(responses [][]byte, mode string) []byte {
	switch mode {
	case "merge":
		return nil
	case "array":
		return nil
	default:
		return responses[0]
	}
}
