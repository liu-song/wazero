//go:build amd64
// +build amd64

package jit

import (
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"
	"github.com/twitchyliquid64/golang-asm/obj/x86"
)

func Test_isIntRegister(t *testing.T) {
	for _, r := range gpIntRegisters {
		require.True(t, isIntRegister(r))
	}
}

func Test_isFloatRegister(t *testing.T) {
	for _, r := range gpFloatRegisters {
		require.True(t, isFloatRegister(r))
	}
}

func TestValueLocationStack_basic(t *testing.T) {
	s := newValueLocationStack()
	// Push stack value.
	loc := s.pushValueOnStack()
	require.Equal(t, uint64(1), s.sp)
	require.Equal(t, uint64(0), loc.stackPointer)
	// Push the register value.
	loc = s.pushValueOnRegister(x86.REG_X1)
	require.Equal(t, uint64(2), s.sp)
	require.Equal(t, uint64(1), loc.stackPointer)
	require.Equal(t, int16(x86.REG_X1), loc.register)
	require.Contains(t, s.usedRegisters, loc.register)
	// markRegisterUsed.
	s.markRegisterUsed(x86.REG_X2)
	require.Contains(t, s.usedRegisters, int16(x86.REG_X2))
	// releaseRegister.
	s.releaseRegister(loc)
	require.NotContains(t, s.usedRegisters, loc.register)
	require.Equal(t, int16(-1), loc.register)
	// Clone.
	cloned := s.clone()
	require.Equal(t, s.usedRegisters, cloned.usedRegisters)
	require.Equal(t, len(s.stack), len(cloned.stack))
	require.Equal(t, s.sp, cloned.sp)
	for i := 0; i < int(s.sp); i++ {
		actual, exp := s.stack[i], cloned.stack[i]
		require.NotEqual(t, uintptr(unsafe.Pointer(exp)), uintptr(unsafe.Pointer(actual)))
	}
}

func TestValueLocationStack_takeFreeRegister(t *testing.T) {
	s := newValueLocationStack()
	// For int registers.
	r, ok := s.takeFreeRegister(gpTypeInt)
	require.True(t, ok)
	require.True(t, isIntRegister(r))
	// Mark all the int registers used.
	for _, r := range gpIntRegisters {
		s.markRegisterUsed(r)
	}
	// Now we cannot take free ones for int.
	_, ok = s.takeFreeRegister(gpTypeInt)
	require.False(t, ok)
	// But we still should be able to take float regs.
	r, ok = s.takeFreeRegister(gpTypeFloat)
	require.True(t, ok)
	require.True(t, isFloatRegister(r))
	// Mark all the float registers used.
	for _, r := range gpFloatRegisters {
		s.markRegisterUsed(r)
	}
	// Now we cannot take free ones for floats.
	_, ok = s.takeFreeRegister(gpTypeFloat)
	require.False(t, ok)
}

func TestValueLocationStack_takeStealTargetFromUsedRegister(t *testing.T) {
	s := newValueLocationStack()
	intReg := int16(x86.REG_R10)
	intLocation := &valueLocation{register: intReg}
	floatReg := int16(x86.REG_X0)
	floatLocation := &valueLocation{register: floatReg}
	s.push(intLocation)
	s.push(floatLocation)
	// Take for float.
	target, ok := s.takeStealTargetFromUsedRegister(gpTypeFloat)
	require.True(t, ok)
	require.Equal(t, floatLocation, target)
	// Take for ints.
	target, ok = s.takeStealTargetFromUsedRegister(gpTypeInt)
	require.True(t, ok)
	require.Equal(t, intLocation, target)
	// Pop float value.
	popped := s.pop()
	require.Equal(t, floatLocation, popped)
	// Now we cannot find the steal target.
	target, ok = s.takeStealTargetFromUsedRegister(gpTypeFloat)
	require.False(t, ok)
	require.Nil(t, target)
	// Pop int value.
	popped = s.pop()
	require.Equal(t, intLocation, popped)
	// Now we cannot find the steal target.
	target, ok = s.takeStealTargetFromUsedRegister(gpTypeInt)
	require.False(t, ok)
	require.Nil(t, target)
}