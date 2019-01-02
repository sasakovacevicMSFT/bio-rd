package packet

import (
	"bytes"
	"fmt"

	"github.com/bio-routing/bio-rd/protocols/isis/types"
	"github.com/bio-routing/bio-rd/util/decode"
	"github.com/bio-routing/bio-rd/util/math"
	"github.com/taktv6/tflow2/convert"
)

const (
	LSPIDLen    = 8
	LSPDUMinLen = 19
	MODX        = 5802
)

// LSPID represents a Link State Packet ID
type LSPID struct {
	SystemID     types.SystemID
	PseudonodeID uint16
}

func (l *LSPID) String() string {
	return fmt.Sprintf("%02d%02d%02d.%02d%02d%02d-%02d", l.SystemID[0], l.SystemID[1], l.SystemID[2], l.SystemID[3], l.SystemID[4], l.SystemID[5], l.PseudonodeID)
}

// Serialize serializes an LSPID
func (l *LSPID) Serialize(buf *bytes.Buffer) {
	buf.Write(l.SystemID[:])
	buf.Write(convert.Uint16Byte(l.PseudonodeID))
}

// Compare returns 1 if l is bigger m, 0 if they are equal, else -1
func (l *LSPID) Compare(m LSPID) int {
	for i := 0; i < 6; i++ {
		if l.SystemID[i] > m.SystemID[i] {
			return 1
		}

		if l.SystemID[i] < m.SystemID[i] {
			return -1
		}
	}

	if l.PseudonodeID > m.PseudonodeID {
		return 1
	}

	if l.PseudonodeID < m.PseudonodeID {
		return -1
	}

	return 0
}

// LSPDU represents a link state PDU
type LSPDU struct {
	Length            uint16
	RemainingLifetime uint16
	LSPID             LSPID
	SequenceNumber    uint32
	Checksum          uint16
	TypeBlock         uint8
	TLVs              []TLV
}

// UpdateLength updates the length of the LSPDU
func (l *LSPDU) updateLength() {
	l.Length = LSPDUMinLen
	for i := range l.TLVs {
		l.Length += uint16(l.TLVs[i].Length())
	}
}

func csum(input []byte) uint16 {
	x := 0
	y := 0
	c0 := 0
	c1 := 0
	partialLen := 0
	i := 0
	left := len(input)

	for left != 0 {
		partialLen = math.Min(left, MODX)

		for i = 0; i < partialLen; i++ {
			c0 = c0 + int(input[i])
			c1 += c0
		}

		c0 = c0 % 255
		c1 = c1 % 255

		left -= partialLen
	}
	fmt.Printf("c0: %d\n", c0)
	fmt.Printf("c1: %d\n", c1)

	fmt.Printf("len(input)-12-1 = %d\n", len(input)-12-1)
	x = ((len(input)-12-1)*c0 - c1) % 255

	if x < 0 {
		x += 255
	}

	y = 510 - c0 - x
	if y > 255 {
		y -= 255
	}

	fmt.Printf("x: %x\n", x)
	fmt.Printf("y: %x\n", y)
	return uint16(x)*256 + uint16(y)
}

func csum2(input []byte) (uint8, uint8) {
	c0 := uint16(0)
	c1 := uint16(0)
	ck0 := uint8(0)
	ck1 := uint8(0)

	for i := range input {
		fmt.Printf("i: %d\n", i)
		c0 = (c0 + uint16(input[i])) % 255
		c1 = (c1 + c0) % 255
		fmt.Printf("c0: %d\n", c0)
		fmt.Printf("c1: %d\n", c1)
	}

	ck0 = uint8((255 - (c0+c1)%255))
	ck1 = uint8(c1 % 255)

	if ck0 == 0 {
		ck0 = 255
	}

	if ck1 == 0 {
		ck1 = 255
	}

	fmt.Printf("CSUM: %x %x\n", ck0, ck1)
	return uint8(ck0), uint8(ck1)
}

// SetChecksum sets the checksum of an LSPDU
func (l *LSPDU) SetChecksum() {
	buf := bytes.NewBuffer(nil)
	l.SerializeChecksumRelevant(buf)
	fmt.Printf("Input: %v\n", buf.Bytes())
	l.Checksum = csum(buf.Bytes())
}

// SerializeChecksumRelevant serializes all fields after the Remaining Lifetime field.
func (l *LSPDU) SerializeChecksumRelevant(buf *bytes.Buffer) {
	l.LSPID.Serialize(buf)
	buf.Write(convert.Uint32Byte(l.SequenceNumber))
	buf.Write(convert.Uint16Byte(l.Checksum))
	buf.WriteByte(l.TypeBlock)

	for _, TLV := range l.TLVs {
		TLV.Serialize(buf)
	}
}

// Serialize serializes a linke state PDU
func (l *LSPDU) Serialize(buf *bytes.Buffer) {
	l.updateLength()
	buf.Write(convert.Uint16Byte(l.Length))
	buf.Write(convert.Uint16Byte(l.RemainingLifetime))
	l.SerializeChecksumRelevant(buf)
}

// DecodeLSPDU decodes an LSPDU
func DecodeLSPDU(buf *bytes.Buffer) (*LSPDU, error) {
	pdu := &LSPDU{}

	fields := []interface{}{
		&pdu.Length,
		&pdu.RemainingLifetime,
		&pdu.LSPID.SystemID,
		&pdu.LSPID.PseudonodeID,
		&pdu.SequenceNumber,
		&pdu.Checksum,
		&pdu.TypeBlock,
	}

	err := decode.Decode(buf, fields)
	if err != nil {
		return nil, fmt.Errorf("Unable to decode fields: %v", err)
	}

	TLVs, err := readTLVs(buf)
	if err != nil {
		return nil, fmt.Errorf("Unable to read TLVs: %v", err)
	}

	pdu.TLVs = TLVs
	return pdu, nil
}
