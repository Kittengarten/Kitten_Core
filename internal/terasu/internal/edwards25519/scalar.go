// Copyright (c) 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package edwards25519

import (
	"errors"

	"github.com/fumiama/terasu/internal/byteorder"
)

// A Scalar is an integer modulo
//
//	l = 2^252 + 27742317777372353535851937790883648493
//
// which is the prime order of the edwards25519 group.
//
// This type works similarly to math/big.Int, and all arguments and
// receivers are allowed to alias.
//
// The zero value is a valid zero element.
type Scalar struct {
	// s is the scalar in the Montgomery domain, in the format of the
	// fiat-crypto implementation.
	s fiatScalarMontgomeryDomainFieldElement
}

// The field implementation in scalar_fiat.go is generated by the fiat-crypto
// project (https://github.com/mit-plv/fiat-crypto) at version v0.0.9 (23d2dbc)
// from a formally verified model.
//
// fiat-crypto code comes under the following license.
//
//     Copyright (c) 2015-2020 The fiat-crypto Authors. All rights reserved.
//
//     Redistribution and use in source and binary forms, with or without
//     modification, are permitted provided that the following conditions are
//     met:
//
//         1. Redistributions of source code must retain the above copyright
//         notice, this list of conditions and the following disclaimer.
//
//     THIS SOFTWARE IS PROVIDED BY the fiat-crypto authors "AS IS"
//     AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO,
//     THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
//     PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL Berkeley Software Design,
//     Inc. BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
//     EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
//     PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
//     PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
//     LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
//     NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
//     SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//

// NewScalar returns a new zero Scalar.
func NewScalar() *Scalar {
	return &Scalar{}
}

// MultiplyAdd sets s = x * y + z mod l, and returns s. It is equivalent to
// using Multiply and then Add.
func (s *Scalar) MultiplyAdd(x, y, z *Scalar) *Scalar {
	// Make a copy of z in case it aliases s.
	zCopy := new(Scalar).Set(z)
	return s.Multiply(x, y).Add(s, zCopy)
}

// Add sets s = x + y mod l, and returns s.
func (s *Scalar) Add(x, y *Scalar) *Scalar {
	// s = 1 * x + y mod l
	fiatScalarAdd(&s.s, &x.s, &y.s)
	return s
}

// Subtract sets s = x - y mod l, and returns s.
func (s *Scalar) Subtract(x, y *Scalar) *Scalar {
	// s = -1 * y + x mod l
	fiatScalarSub(&s.s, &x.s, &y.s)
	return s
}

// Negate sets s = -x mod l, and returns s.
func (s *Scalar) Negate(x *Scalar) *Scalar {
	// s = -1 * x + 0 mod l
	fiatScalarOpp(&s.s, &x.s)
	return s
}

// Multiply sets s = x * y mod l, and returns s.
func (s *Scalar) Multiply(x, y *Scalar) *Scalar {
	// s = x * y + 0 mod l
	fiatScalarMul(&s.s, &x.s, &y.s)
	return s
}

// Set sets s = x, and returns s.
func (s *Scalar) Set(x *Scalar) *Scalar {
	*s = *x
	return s
}

// SetUniformBytes sets s = x mod l, where x is a 64-byte little-endian integer.
// If x is not of the right length, SetUniformBytes returns nil and an error,
// and the receiver is unchanged.
//
// SetUniformBytes can be used to set s to a uniformly distributed value given
// 64 uniformly distributed random bytes.
func (s *Scalar) SetUniformBytes(x []byte) (*Scalar, error) {
	if len(x) != 64 {
		return nil, errors.New("edwards25519: invalid SetUniformBytes input length")
	}

	// We have a value x of 512 bits, but our fiatScalarFromBytes function
	// expects an input lower than l, which is a little over 252 bits.
	//
	// Instead of writing a reduction function that operates on wider inputs, we
	// can interpret x as the sum of three shorter values a, b, and c.
	//
	//    x = a + b * 2^168 + c * 2^336  mod l
	//
	// We then precompute 2^168 and 2^336 modulo l, and perform the reduction
	// with two multiplications and two additions.

	s.setShortBytes(x[:21])
	t := new(Scalar).setShortBytes(x[21:42])
	s.Add(s, t.Multiply(t, scalarTwo168))
	t.setShortBytes(x[42:])
	s.Add(s, t.Multiply(t, scalarTwo336))

	return s, nil
}

// scalarTwo168 and scalarTwo336 are 2^168 and 2^336 modulo l, encoded as a
// fiatScalarMontgomeryDomainFieldElement, which is a little-endian 4-limb value
// in the 2^256 Montgomery domain.
var scalarTwo168 = &Scalar{s: [4]uint64{
	0x5b8ab432eac74798, 0x38afddd6de59d5d7,
	0xa2c131b399411b7c, 0x6329a7ed9ce5a30,
}}

var scalarTwo336 = &Scalar{s: [4]uint64{
	0xbd3d108e2b35ecc5, 0x5c3a3718bdf9c90b,
	0x63aa97a331b4f2ee, 0x3d217f5be65cb5c,
}}

// setShortBytes sets s = x mod l, where x is a little-endian integer shorter
// than 32 bytes.
func (s *Scalar) setShortBytes(x []byte) *Scalar {
	if len(x) >= 32 {
		panic("edwards25519: internal error: setShortBytes called with a long string")
	}
	var buf [32]byte
	copy(buf[:], x)
	fiatScalarFromBytes((*[4]uint64)(&s.s), &buf)
	fiatScalarToMontgomery(&s.s, (*fiatScalarNonMontgomeryDomainFieldElement)(&s.s))
	return s
}

// SetCanonicalBytes sets s = x, where x is a 32-byte little-endian encoding of
// s, and returns s. If x is not a canonical encoding of s, SetCanonicalBytes
// returns nil and an error, and the receiver is unchanged.
func (s *Scalar) SetCanonicalBytes(x []byte) (*Scalar, error) {
	if len(x) != 32 {
		return nil, errors.New("invalid scalar length")
	}
	if !isReduced(x) {
		return nil, errors.New("invalid scalar encoding")
	}

	fiatScalarFromBytes((*[4]uint64)(&s.s), (*[32]byte)(x))
	fiatScalarToMontgomery(&s.s, (*fiatScalarNonMontgomeryDomainFieldElement)(&s.s))

	return s, nil
}

// scalarMinusOneBytes is l - 1 in little endian.
var scalarMinusOneBytes = [32]byte{236, 211, 245, 92, 26, 99, 18, 88, 214, 156, 247, 162, 222, 249, 222, 20, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 16}

// isReduced returns whether the given scalar in 32-byte little endian encoded
// form is reduced modulo l.
func isReduced(s []byte) bool {
	if len(s) != 32 {
		return false
	}

	for i := len(s) - 1; i >= 0; i-- {
		switch {
		case s[i] > scalarMinusOneBytes[i]:
			return false
		case s[i] < scalarMinusOneBytes[i]:
			return true
		}
	}
	return true
}

// SetBytesWithClamping applies the buffer pruning described in RFC 8032,
// Section 5.1.5 (also known as clamping) and sets s to the result. The input
// must be 32 bytes, and it is not modified. If x is not of the right length,
// SetBytesWithClamping returns nil and an error, and the receiver is unchanged.
//
// Note that since Scalar values are always reduced modulo the prime order of
// the curve, the resulting value will not preserve any of the cofactor-clearing
// properties that clamping is meant to provide. It will however work as
// expected as long as it is applied to points on the prime order subgroup, like
// in Ed25519. In fact, it is lost to history why RFC 8032 adopted the
// irrelevant RFC 7748 clamping, but it is now required for compatibility.
func (s *Scalar) SetBytesWithClamping(x []byte) (*Scalar, error) {
	// The description above omits the purpose of the high bits of the clamping
	// for brevity, but those are also lost to reductions, and are also
	// irrelevant to edwards25519 as they protect against a specific
	// implementation bug that was once observed in a generic Montgomery ladder.
	if len(x) != 32 {
		return nil, errors.New("edwards25519: invalid SetBytesWithClamping input length")
	}

	// We need to use the wide reduction from SetUniformBytes, since clamping
	// sets the 2^254 bit, making the value higher than the order.
	var wideBytes [64]byte
	copy(wideBytes[:], x[:])
	wideBytes[0] &= 248
	wideBytes[31] &= 63
	wideBytes[31] |= 64
	return s.SetUniformBytes(wideBytes[:])
}

// Bytes returns the canonical 32-byte little-endian encoding of s.
func (s *Scalar) Bytes() []byte {
	// This function is outlined to make the allocations inline in the caller
	// rather than happen on the heap.
	var encoded [32]byte
	return s.bytes(&encoded)
}

func (s *Scalar) bytes(out *[32]byte) []byte {
	var ss fiatScalarNonMontgomeryDomainFieldElement
	fiatScalarFromMontgomery(&ss, &s.s)
	fiatScalarToBytes(out, (*[4]uint64)(&ss))
	return out[:]
}

// Equal returns 1 if s and t are equal, and 0 otherwise.
func (s *Scalar) Equal(t *Scalar) int {
	var diff fiatScalarMontgomeryDomainFieldElement
	fiatScalarSub(&diff, &s.s, &t.s)
	var nonzero uint64
	fiatScalarNonzero(&nonzero, (*[4]uint64)(&diff))
	nonzero |= nonzero >> 32
	nonzero |= nonzero >> 16
	nonzero |= nonzero >> 8
	nonzero |= nonzero >> 4
	nonzero |= nonzero >> 2
	nonzero |= nonzero >> 1
	return int(^nonzero) & 1
}

// nonAdjacentForm computes a width-w non-adjacent form for this scalar.
//
// w must be between 2 and 8, or nonAdjacentForm will panic.
func (s *Scalar) nonAdjacentForm(w uint) [256]int8 {
	// This implementation is adapted from the one
	// in curve25519-dalek and is documented there:
	// https://github.com/dalek-cryptography/curve25519-dalek/blob/f630041af28e9a405255f98a8a93adca18e4315b/src/scalar.rs#L800-L871
	b := s.Bytes()
	if b[31] > 127 {
		panic("scalar has high bit set illegally")
	}
	if w < 2 {
		panic("w must be at least 2 by the definition of NAF")
	} else if w > 8 {
		panic("NAF digits must fit in int8")
	}

	var naf [256]int8
	var digits [5]uint64

	for i := 0; i < 4; i++ {
		digits[i] = byteorder.LeUint64(b[i*8:])
	}

	width := uint64(1 << w)
	windowMask := uint64(width - 1)

	pos := uint(0)
	carry := uint64(0)
	for pos < 256 {
		indexU64 := pos / 64
		indexBit := pos % 64
		var bitBuf uint64
		if indexBit < 64-w {
			// This window's bits are contained in a single u64
			bitBuf = digits[indexU64] >> indexBit
		} else {
			// Combine the current 64 bits with bits from the next 64
			bitBuf = (digits[indexU64] >> indexBit) | (digits[1+indexU64] << (64 - indexBit))
		}

		// Add carry into the current window
		window := carry + (bitBuf & windowMask)

		if window&1 == 0 {
			// If the window value is even, preserve the carry and continue.
			// Why is the carry preserved?
			// If carry == 0 and window & 1 == 0,
			//    then the next carry should be 0
			// If carry == 1 and window & 1 == 0,
			//    then bit_buf & 1 == 1 so the next carry should be 1
			pos += 1
			continue
		}

		if window < width/2 {
			carry = 0
			naf[pos] = int8(window)
		} else {
			carry = 1
			naf[pos] = int8(window) - int8(width)
		}

		pos += w
	}
	return naf
}

func (s *Scalar) signedRadix16() [64]int8 {
	b := s.Bytes()
	if b[31] > 127 {
		panic("scalar has high bit set illegally")
	}

	var digits [64]int8

	// Compute unsigned radix-16 digits:
	for i := 0; i < 32; i++ {
		digits[2*i] = int8(b[i] & 15)
		digits[2*i+1] = int8((b[i] >> 4) & 15)
	}

	// Recenter coefficients:
	for i := 0; i < 63; i++ {
		carry := (digits[i] + 8) >> 4
		digits[i] -= carry << 4
		digits[i+1] += carry
	}

	return digits
}
