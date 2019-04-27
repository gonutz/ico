package ico

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"

	"github.com/nfnt/resize"
)

func FromImage(img image.Image) []byte {
	var buf bytes.Buffer
	buf.Write([]byte{
		0, 0, // reserved
		1, 0, // 1 for .ico, 2 for .cur
		9, 0, // number of images in this icon
	})
	sizes := []int{16, 24, 32, 48, 64, 96, 128, 192, 256}
	var images [][]byte
	for _, size := range sizes {
		images = append(images, resizedImageData(img, size))
	}
	fileOffset := 6 + len(sizes)*16 // header size, first image starts here
	for i, size := range sizes {
		buf.Write([]byte{
			// width,height; 256 wraps to 0 which is interpreted as 256
			byte(size), byte(size),
			0,    // no color palette
			0,    // reserved
			1, 0, // color planes, unused
			32, 0, // bits per pixel
		})
		// append data length
		binary.Write(&buf, binary.LittleEndian, uint32(len(images[i])))
		// append data start offset in bytes from file start
		binary.Write(&buf, binary.LittleEndian, uint32(fileOffset))
		fileOffset += len(images[i]) // next image starts after this one
	}
	for _, img := range images {
		buf.Write(img)
	}
	return buf.Bytes()
}

func resizedImageData(img image.Image, size int) []byte {
	if img.Bounds().Dx() != size || img.Bounds().Dy() != size {
		img = resize.Resize(uint(size), uint(size), img, resize.MitchellNetravali)
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}
