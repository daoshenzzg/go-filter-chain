package filterchain

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRunFilters(t *testing.T) {
	chain := Chain{}

	results := []int{}

	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 1)
		err := chain.Next(args...)
		results = append(results, -1)
		return err
	}})
	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 2)
		err := chain.Next(args...)
		results = append(results, -2)
		return err
	}})
	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 3)
		err := chain.Next(args...)
		results = append(results, -3)
		return err
	}})

	chain.Execute()

	assert.Equal(t, 1, results[0])
	assert.Equal(t, 2, results[1])
	assert.Equal(t, 3, results[2])
	assert.Equal(t, -3, results[3])
	assert.Equal(t, -2, results[4])
	assert.Equal(t, -1, results[5])
}

func TestStopRunningOnError(t *testing.T) {
	chain := Chain{}

	results := []int{}

	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 1)
		err := chain.Next()
		results = append(results, -1)
		return err
	}})
	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 2)
		return errors.New("Error!")
	}})
	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 3)
		err := chain.Next()
		results = append(results, -3)
		return err
	}})

	chain.Execute()

	assert.Equal(t, 3, len(results))

	assert.Equal(t, 1, results[0])
	assert.Equal(t, 2, results[1])
	assert.Equal(t, -1, results[2])
}

func TestPropagateError(t *testing.T) {
	chain := Chain{}

	results := []int{}

	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 1)
		err := chain.Next()
		results = append(results, -1)
		return err
	}})
	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 2)
		return errors.New("Error!")
	}})
	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 3)
		err := chain.Next()
		results = append(results, -3)
		return err
	}})

	err := chain.Execute()

	assert.NotNil(t, err)
	assert.Equal(t, "Error!", err.Error())
}

func TestNotRunAgain(t *testing.T) {
	chain := Chain{}

	results := []int{}

	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 1)
		chain.Next()
		results = append(results, -1)
		return nil
	}})

	chain.Execute()
	chain.Execute()
	chain.Execute()

	assert.Equal(t, 2, len(results))
}

func TestRewindChain(t *testing.T) {
	chain := Chain{}

	results := []int{}

	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 1)
		chain.Next()
		results = append(results, -1)
		return nil
	}})

	chain.Execute()
	chain.Rewind()
	chain.Execute()
	chain.Rewind()
	chain.Execute()
	chain.Rewind()

	assert.Equal(t, 6, len(results))
}

type CustomFilter struct {
	run int
}

func (filter *CustomFilter) Execute(chain *Chain, args ...interface{}) error {
	filter.run++
	err := chain.Next(args...)
	return err
}

func TestCustomStruct(t *testing.T) {
	chain := Chain{}

	filter := &CustomFilter{}

	chain.AddFilter(filter)

	chain.Execute()

	assert.Equal(t, 1, filter.run)
}

func TestMethodChaining(t *testing.T) {
	chain := Chain{}

	results := []int{}

	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 1)
		err := chain.Next()
		results = append(results, -1)
		return err
	}}).AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
		results = append(results, 2)
		err := chain.Next()
		results = append(results, -2)
		return err
	}})

	chain.Execute()

	assert.Equal(t, 4, len(results))
}

func TestParamsPassing(t *testing.T) {
	chain := Chain{}

	results := []string{}

	chain.AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
        foo := args[0].(string)
        foo = foo + "bar"
		err := chain.Next(foo)
		results = append(results, foo)
		return err
	}}).AddFilter(&Inline{func(chain *Chain, args ...interface{}) error {
        foo := args[0].(string)
        foo = foo + "baz"
		err := chain.Next(foo)
		results = append(results, foo)
		return err
	}})

	chain.Execute("foo")

	assert.Equal(t, 2, len(results))
	assert.Equal(t, "foobarbaz", results[0])
	assert.Equal(t, "foobar", results[1])
}
