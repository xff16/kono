package bravka

import "go.uber.org/zap"

type aggregator interface {
	aggregate(responses [][]byte, mode string) []byte
}

type defaultAggregator struct {
	log *zap.Logger
}

func (a *defaultAggregator) aggregate(responses [][]byte, mode string) []byte {
	var result []byte

	switch mode {
	case "merge":
		return nil
	case "array":
		for _, response := range responses {
			result = append(result, response...)
		}
		return nil
	default:
		return responses[0]
	}
}
