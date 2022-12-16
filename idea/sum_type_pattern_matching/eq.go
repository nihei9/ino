package main

type Eqer interface {
	Eq(Eqer) bool
}

type B bool
type U8 uint8
type U16 uint16
type U32 uint32
type U64 uint64
type I8 int8
type I16 int16
type I32 int32
type I64 int64
type F32 float32
type F64 float64
type C64 complex64
type C128 complex128
type S string
type I int
type U uint
type P uintptr
type Byte byte
type R rune

func (x B) Eq(y Eqer) bool    { return x == y }
func (x U8) Eq(y Eqer) bool   { return x == y }
func (x U16) Eq(y Eqer) bool  { return x == y }
func (x U32) Eq(y Eqer) bool  { return x == y }
func (x U64) Eq(y Eqer) bool  { return x == y }
func (x I8) Eq(y Eqer) bool   { return x == y }
func (x I16) Eq(y Eqer) bool  { return x == y }
func (x I32) Eq(y Eqer) bool  { return x == y }
func (x I64) Eq(y Eqer) bool  { return x == y }
func (x F32) Eq(y Eqer) bool  { return x == y }
func (x F64) Eq(y Eqer) bool  { return x == y }
func (x C64) Eq(y Eqer) bool  { return x == y }
func (x C128) Eq(y Eqer) bool { return x == y }
func (x S) Eq(y Eqer) bool    { return x == y }
func (x I) Eq(y Eqer) bool    { return x == y }
func (x U) Eq(y Eqer) bool    { return x == y }
func (x P) Eq(y Eqer) bool    { return x == y }
func (x Byte) Eq(y Eqer) bool { return x == y }
func (x R) Eq(y Eqer) bool    { return x == y }
