package main

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

const (
	ddsHeaderSize = 0x80
	rMask         = uint32(0x00ff0000)
	gMask         = uint32(0x0000ff00)
	bMask         = uint32(0x000000ff)
	aMask         = uint32(0xff000000)
)

type mipLayout struct {
	offset int
	width  uint32
	height uint32
	size   int
}

type ddsLayout struct {
	width  uint32
	height uint32
	masks  [4]uint32
	mips   []mipLayout
}

func parseDDS(input []byte) (ddsLayout, error) {
	if len(input) < ddsHeaderSize {
		return ddsLayout{}, fmt.Errorf("DDS is shorter than 128-byte header")
	}
	if string(input[:4]) != "DDS " {
		return ddsLayout{}, fmt.Errorf("invalid DDS magic")
	}
	u32 := func(offset int) uint32 { return binary.LittleEndian.Uint32(input[offset : offset+4]) }
	if u32(4) != 0x7c {
		return ddsLayout{}, fmt.Errorf("DDS header size is %#x, want 0x7c", u32(4))
	}
	if u32(8)&0x2100f != 0x2100f {
		return ddsLayout{}, fmt.Errorf("DDS header flags %#x omit a required field", u32(8))
	}
	if u32(76) != 0x20 {
		return ddsLayout{}, fmt.Errorf("DDS pixel-format size is %#x, want 0x20", u32(76))
	}
	formatFlags := u32(80)
	if formatFlags&0x41 != 0x41 || formatFlags&0x4 != 0 {
		return ddsLayout{}, fmt.Errorf("DDS pixel-format flags %#x are not uncompressed RGB with alpha", formatFlags)
	}
	if u32(88) != 32 {
		return ddsLayout{}, fmt.Errorf("DDS bit count is %d, want 32", u32(88))
	}
	masks := [4]uint32{u32(92), u32(96), u32(100), u32(104)}
	wantMasks := [4]uint32{rMask, gMask, bMask, aMask}
	if masks != wantMasks {
		return ddsLayout{}, fmt.Errorf("DDS masks are R=%#08x G=%#08x B=%#08x A=%#08x", masks[0], masks[1], masks[2], masks[3])
	}
	width, height, mipCount := u32(16), u32(12), u32(28)
	if width == 0 || height == 0 || mipCount == 0 {
		return ddsLayout{}, fmt.Errorf("DDS dimensions and mip count must be nonzero")
	}
	if uint64(u32(20)) != uint64(width)*4 {
		return ddsLayout{}, fmt.Errorf("DDS pitch is %d, want %d", u32(20), uint64(width)*4)
	}
	if u32(84) != 0 {
		return ddsLayout{}, fmt.Errorf("DDS FourCC must be zero for uncompressed RGB")
	}
	if u32(108)&0x401008 != 0x401008 {
		return ddsLayout{}, fmt.Errorf("DDS caps %#x omit texture, complex, or mipmap", u32(108))
	}
	maxMips := uint32(bits.Len32(max(width, height)))
	if mipCount > maxMips {
		return ddsLayout{}, fmt.Errorf("DDS mip count %d exceeds full chain %d", mipCount, maxMips)
	}

	layout := ddsLayout{width: width, height: height, masks: masks, mips: make([]mipLayout, 0, mipCount)}
	offset := ddsHeaderSize
	mipWidth, mipHeight := width, height
	for level := uint32(0); level < mipCount; level++ {
		size64 := uint64(mipWidth) * uint64(mipHeight) * 4
		if size64 > uint64(len(input)) {
			return ddsLayout{}, fmt.Errorf("mip %d byte size exceeds input", level)
		}
		size := int(size64)
		if size > len(input)-offset {
			return ddsLayout{}, fmt.Errorf("mip %d exceeds DDS payload", level)
		}
		layout.mips = append(layout.mips, mipLayout{offset: offset, width: mipWidth, height: mipHeight, size: size})
		offset += size
		mipWidth = max(1, mipWidth/2)
		mipHeight = max(1, mipHeight/2)
	}
	if offset != len(input) {
		return ddsLayout{}, fmt.Errorf("DDS has %d unexplained trailing bytes", len(input)-offset)
	}
	return layout, nil
}

func (layout ddsLayout) pixelCount() int {
	total := 0
	for _, mip := range layout.mips {
		total += mip.size / 4
	}
	return total
}

func recolorDDS(input []byte) ([]byte, ddsLayout, error) {
	layout, err := parseDDS(input)
	if err != nil {
		return nil, ddsLayout{}, err
	}
	output := append([]byte(nil), input...)
	for _, mip := range layout.mips {
		for offset := mip.offset; offset < mip.offset+mip.size; offset += 4 {
			pixel := binary.LittleEndian.Uint32(input[offset : offset+4])
			r := channel(pixel, layout.masks[0])
			g := channel(pixel, layout.masks[1])
			b := channel(pixel, layout.masks[2])
			r = (r + 255) / 2
			g /= 3
			b = (b + 255) / 2
			pixel = setChannel(pixel, layout.masks[0], r)
			pixel = setChannel(pixel, layout.masks[1], g)
			pixel = setChannel(pixel, layout.masks[2], b)
			binary.LittleEndian.PutUint32(output[offset:offset+4], pixel)
		}
	}
	return output, layout, nil
}

func channel(pixel, mask uint32) uint32 {
	return (pixel & mask) >> bits.TrailingZeros32(mask)
}

func setChannel(pixel, mask, value uint32) uint32 {
	shift := bits.TrailingZeros32(mask)
	return (pixel &^ mask) | ((value << shift) & mask)
}
