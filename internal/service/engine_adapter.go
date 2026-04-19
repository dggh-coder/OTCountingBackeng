package service

import "otcountingbackend/internal/engine"

type EngineAdapter struct{}

func (EngineAdapter) Calculate(input engine.Input) (engine.Output, error) {
	return engine.Calculate(input)
}
