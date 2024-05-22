package transformers

import (
	"context"
	"fmt"

	"github.com/greenmaskio/greenmask/internal/generators"
	int_utils "github.com/greenmaskio/greenmask/internal/generators/transformers/utils"
)

type Int64Limiter struct {
	MinValue int64
	MaxValue int64
	distance uint64
}

// NewInt64Limiter - create limiter by int size. size is required for optional min and max.
func NewInt64Limiter(minValue, maxValue int64, size int) (*Int64Limiter, error) {
	if minValue == 0 || maxValue == 0 {
		minThreshold, maxThreshold, err := int_utils.GetIntThresholds(size)
		if err != nil {
			return nil, err
		}

		if minValue == 0 {
			minValue = minThreshold
		}

		if maxValue == 0 {
			maxValue = maxThreshold
		}
	}

	if minValue >= maxValue {
		return nil, int_utils.ErrWrongLimits
	}

	return &Int64Limiter{
		MinValue: minValue,
		MaxValue: maxValue,
		distance: uint64(maxValue - minValue),
	}, nil
}

func (l *Int64Limiter) Limit(v uint64) int64 {
	res := l.MinValue + int64(v%l.distance)
	if res < 0 {
		return res % l.MinValue
	}
	return res % l.MaxValue
}

type RandomInt64Transformer struct {
	generator  generators.Generator
	limiter    *Int64Limiter
	byteLength int
}

func NewRandomInt64Transformer(limiter *Int64Limiter, size int) (*RandomInt64Transformer, error) {
	return &RandomInt64Transformer{
		limiter:    limiter,
		byteLength: size,
	}, nil
}

func (ig *RandomInt64Transformer) Transform(ctx context.Context, original []byte) (int64, error) {
	var res int64
	var limiter = ig.limiter
	limiterAny := ctx.Value("limiter")

	if limiterAny != nil {
		limiter = limiterAny.(*Int64Limiter)
	}

	resBytes, err := ig.generator.Generate(original)
	if err != nil {
		return 0, err
	}

	if limiter != nil {
		res = limiter.Limit(generators.BuildUint64FromBytes(resBytes))
	} else {
		res = generators.BuildInt64FromBytes(resBytes)
	}

	return res, nil
}

func (ig *RandomInt64Transformer) GetRequiredGeneratorByteLength() int {
	return ig.byteLength
}

func (ig *RandomInt64Transformer) SetGenerator(g generators.Generator) error {
	if g.Size() < ig.byteLength {
		return fmt.Errorf("requested byte length (%d) higher than generator can produce (%d)", ig.byteLength, g.Size())
	}
	ig.generator = g
	return nil
}
