package delay

import (
	"testing"
	"time"
)

func TestDelaySetAndGet(t *testing.T) {
	initialValue := 1000 * time.Millisecond
	modifiedValue := 2000 * time.Millisecond
	deviation := 1000 * time.Millisecond

	fixed := Fixed(initialValue)
	variableNormal := VariableNormal(initialValue, deviation, nil)
	variableUniform := VariableUniform(initialValue, deviation, nil)

	if fixed.Get().Seconds() != 1 {
		t.Fatal("Fixed delay not initialized correctly")
	}

	if variableNormal.Get().Seconds() != 1 {
		t.Fatal("Normalized variable delay not initialized correctly")
	}

	if variableUniform.Get().Seconds() != 1 {
		t.Fatal("Uniform variable delay not initialized correctly")
	}

	fixed.Set(modifiedValue)

	if fixed.Get().Seconds() != 2 {
		t.Fatal("Fixed delay not set correctly")
	}

	variableNormal.Set(modifiedValue)

	if variableNormal.Get().Seconds() != 2 {
		t.Fatal("Normalized variable delay not set correctly")
	}

	variableUniform.Set(modifiedValue)

	if variableUniform.Get().Seconds() != 2 {
		t.Fatal("Uniform variable delay not initialized correctly")
	}

}

type fixedAdd struct {
	toAdd time.Duration
}

func (fa *fixedAdd) NextWaitTime(t time.Duration) time.Duration {
	return t + fa.toAdd
}

func TestDelaySleep(t *testing.T) {
	initialValue := 1000 * time.Millisecond
	toAdd := 500 * time.Millisecond
	generator := &fixedAdd{toAdd: toAdd}
	delay := Delay(initialValue, generator)

	if delay.NextWaitTime() != initialValue+toAdd {
		t.Fatal("NextWaitTime should call the generator")
	}

}
