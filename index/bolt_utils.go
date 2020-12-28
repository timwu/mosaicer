package index

import (
	"encoding/binary"
	"image"
	"math"
)

func pointToBytes(point image.Point) []byte {
	bytes := make([]byte, binary.MaxVarintLen64*2)
	xSize := binary.PutVarint(bytes, int64(point.X))
	ySize := binary.PutVarint(bytes[xSize:], int64(point.Y))
	return bytes[:xSize+ySize]
}

func bytesToPoint(bytes []byte) image.Point {
	x, xSize := binary.Varint(bytes)
	y, _ := binary.Varint(bytes[xSize:])
	return image.Point{X: int(x), Y: int(y)}
}

func intToBytes(i int) []byte {
	bytes := make([]byte, binary.MaxVarintLen64)
	size := binary.PutVarint(bytes, int64(i))
	return bytes[:size]
}

func bytesToInt(bytes []byte) int {
	i, _ := binary.Varint(bytes)
	return int(i)
}

func floatsToBytes(floats []float64) []byte {
	buf := make([]byte, len(floats)*8)
	for i, f := range floats {
		binary.BigEndian.PutUint64(buf[i*8:], math.Float64bits(f))
	}
	return buf
}

func bytesToFloats(in []byte) []float64 {
	floats := make([]float64, len(in)/8)
	for i := range floats {
		floats[i] = math.Float64frombits(binary.BigEndian.Uint64(in[i*8:]))
	}
	return floats
}
