package util

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type UtilSuite struct {
	suite.Suite
}

func (s *UtilSuite) TestMin() {
	assert := s.Require()

	assert.Equal(1, MinInt(1, 2))
	assert.Equal(1, MinInt(4, 1))
}

func (s *UtilSuite) TestMax() {
	assert := s.Require()

	assert.Equal(2, MaxInt(1, 2))
	assert.Equal(4, MaxInt(4, 1))
}

func (s *UtilSuite) TestClamp() {
	assert := s.Require()

	assert.Equal(0, ClampInt(-1, 0, 10))
	assert.Equal(1, ClampInt(1, 0, 10))
	assert.Equal(10, ClampInt(12, 0, 10))
}

func (s *UtilSuite) TestCeilDiv() {
	assert := s.Require()

	assert.Equal(2, CeilDiv(4, 2))
	assert.Equal(3, CeilDiv(5, 2))
}

func (s *UtilSuite) TestEnvironment() {
	assert := s.Require()

	parent := NewEnvironment(nil)
	parent.Set("var1", "1")

	env1 := NewEnvironment(parent)
	env1.Set("var2", "2")

	env2 := NewEnvironment(parent)
	env2.Set("var3", "3")

	env3 := NewEnvironment(env2)
	env3.Set("var4", "4")

	assert.True(parent.Exists("var1"))
	assert.False(parent.Exists("var2"))

	assert.True(env1.Exists("var1"))
	assert.True(env1.Exists("var2"))
	assert.False(env1.Exists("var3"))

	assert.True(env2.Exists("var1"))
	assert.False(env2.Exists("var2"))
	assert.True(env2.Exists("var3"))

	assert.True(env3.Exists("var1"))
	assert.False(env3.Exists("var2"))
	assert.True(env3.Exists("var3"))
	assert.True(env3.Exists("var4"))

	assert.Equal("1", env3.Get("var1"))

	// Test string case
	assert.Equal("1", env3.Get("Var1"))
	assert.Equal("1", env3.Get("vAr1"))
	assert.Equal("1", env3.Get("VaR1"))

	csEnv3 := NewEnvironmentCaseSensitive(nil)
	csEnv3.SetFromMap(env3.ToMap())

	assert.Equal("1", csEnv3.Get("var1"))
	assert.Empty(csEnv3.Get("Var1"))
	assert.Empty(csEnv3.Get("vAr1"))
	assert.Empty(csEnv3.Get("VAR1"))

	assert.Empty(env3.Get("var2"))
	assert.Equal("3", env3.Get("var3"))
	assert.Equal("4", env3.Get("var4"))

	env3.Set("var5", "5")
	assert.Equal("5", env3.Get("var5"))

	parent.Set("var0", "0")
	assert.Equal("0", env3.Get("var0"))

	parent.Clear()
	assert.Empty(env3.Get("var0"))
	assert.Empty(env3.Get("var1"))
}

func (s *UtilSuite) TestEscapeStringVariables() {
	assert := s.Require()

	variables := map[string]string{
		"v1": "1",
		"v2": "two",
		"v3": "3",
	}

	env := NewEnvironment(nil)
	env.SetFromMap(variables)

	res, err := env.EscapeStringVariables("{{v1}")
	assert.ErrorIs(err, ErrUnmatchedLeftBrace)
	assert.Empty(res)

	res, err = env.EscapeStringVariables("{v1}}")
	assert.ErrorIs(err, ErrUnmatchedRightBrace)
	assert.Empty(res)

	res, err = env.EscapeStringVariables("{v1")
	assert.ErrorIs(err, ErrUnmatchedLeftBrace)
	assert.Empty(res)

	res, err = env.EscapeStringVariables("v1}")
	assert.ErrorIs(err, ErrUnmatchedRightBrace)
	assert.Empty(res)

	res, err = env.EscapeStringVariables("{v4}")
	assert.ErrorIs(err, ErrVariableNotFound)
	assert.Empty(res)

	res, err = env.EscapeStringVariables("{v1} {v2} {v3} {v4}")
	assert.ErrorIs(err, ErrVariableNotFound)
	assert.Empty(res)

	res, err = env.EscapeStringVariables("{v1}")
	assert.NoError(err)
	assert.Equal(env.Get("v1"), res)

	expected := env.Get("v1") + " -> " + env.Get("v2")
	res, err = env.EscapeStringVariables("{v1} -> {v2}")
	assert.NoError(err)
	assert.Equal(expected, res)
}

func TestUtil(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}
