package util

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

type SetTestSuite struct {
	suite.Suite
}

func (s *SetTestSuite) TestCreate() {
	assert := s.Require()

	setA := NewSet()
	setA.Add(1)
	assert.Equal(1, setA.Size())

	setA.AddMulti(2, 3, 4)
	assert.Equal(4, setA.Size())

	assert.True(setA.Has(1))
	assert.False(setA.Has(10))

	setA.Remove(1)
	assert.Equal(3, setA.Size())
	assert.False(setA.Has(1))

	setA.Clear()
	assert.Zero(setA.Size())
}

func (s *SetTestSuite) TestEqual() {
	assert := s.Require()

	setA := NewSet(1, 2, 3)
	setB := NewSet(1, 2)
	setC := NewSet(1, 2, 4)
	setD := NewSet(3, 2, 1)
	setE := NewSet(1, 3, 2)

	assert.False(setA.Equal(setB))
	assert.False(setA.Equal(setC))
	assert.True(setA.Equal(setD))
	assert.True(setA.Equal(setE))
}

func (s *SetTestSuite) TestFilter() {
	assert := s.Require()

	setA := NewSet(1, 2, 3, 4, 5, 6)
	expected := NewSet(2, 4, 6)

	setB := setA.Filter(func(v interface{}) bool {
		intV := v.(int)
		return intV%2 == 0
	})

	assert.Equal(expected.Size(), setB.Size())
	assert.True(expected.Equal(setB))
}

func (s *SetTestSuite) TestIntersection() {
	assert := s.Require()

	setA := NewSet(1, 2, 3, 4, 5, 6)
	setB := NewSet(2, 4, 6, 1)
	expected := NewSet(1, 2, 4, 6)

	setC := setA.Intersect(setB)

	assert.Equal(expected.Size(), setC.Size())
	assert.True(expected.Equal(setC))
}

func (s *SetTestSuite) TestUnion() {
	assert := s.Require()

	setA := NewSet(1, 2, 3, 4, 5)
	setB := NewSet(6, 7, 8, 9)

	expected := NewSet(9, 8, 7, 6, 5, 4, 3, 2, 1)

	setC := setA.Union(setB)
	assert.Equal(expected.Size(), setC.Size())
	assert.True(expected.Equal(setC))
}

func (s *SetTestSuite) TestDifference() {
	assert := s.Require()

	setA := NewSet(1, 2, 3, 4)
	setB := NewSet(1, 4)
	expected := NewSet(2, 3)

	diff := setA.Difference(setB)
	assert.Equal(expected.Size(), diff.Size())
	assert.True(expected.Equal(diff))
}

func TestSet(t *testing.T) {
	suite.Run(t, new(SetTestSuite))
}
