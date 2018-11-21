package delay

import (
	"math"
	"math/rand"
	"testing"
	"time"
)

const testSeed = 99

func TestGeneratorNextWaitTime(t *testing.T) {
	initialValue := 1000 * time.Millisecond
	deviation := 1000 * time.Millisecond

	firstRandomNormal := rand.New(rand.NewSource(testSeed)).NormFloat64()
	firstRandom := rand.New(rand.NewSource(testSeed)).Float64()
	fixedGenerator := FixedGenerator()
	variableNormalGenerator := VariableNormalGenerator(deviation, rand.New(rand.NewSource(testSeed)))
	variableUniformGenerator := VariableUniformGenerator(deviation, rand.New(rand.NewSource(testSeed)))

	if fixedGenerator.NextWaitTime(initialValue).Seconds() != 1 {
		t.Fatal("Fixed generator output incorrect wait time")
	}

	if math.Abs(variableNormalGenerator.NextWaitTime(initialValue).Seconds()-(firstRandomNormal*deviation.Seconds()+initialValue.Seconds())) > 0.00001 {
		t.Fatal("Normalized variable delay generator output incorrect wait time")
	}

	if math.Abs(variableUniformGenerator.NextWaitTime(initialValue).Seconds()-(firstRandom*deviation.Seconds()+initialValue.Seconds())) > 0.00001 {
		t.Fatal("Uniform variable delay output incorrect wait time")
	}

}
