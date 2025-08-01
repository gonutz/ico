package ico

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"os"

	_ "github.com/gonutz/bmp"
	_ "image/gif"
	_ "image/jpeg"
)

func FromFile(path string) ([]byte, error) {
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"ico.FromFile: failed to read file '%s': %w",
			path, err,
		)
	}

	if bytes.HasPrefix(data, pngHeader) {
		// Use PNG file data as is without re-encoding it.
		config, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf(
				"ico.FromFile: failed to decode PNG header of file '%s': %w",
				path, err,
			)
		}
		return pngToIco(
			fmt.Sprintf("ico.FromFile: failed to convert '%s'", path),
			config.Width,
			config.Height,
			data,
		)
	} else {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf(
				"ico.FromFile: failed to decode image file '%s': %w",
				path, err,
			)
		}

		var pngBuf bytes.Buffer
		if err := png.Encode(&pngBuf, img); err != nil {
			return nil, err
		}
		pngData := pngBuf.Bytes()

		return pngToIco(
			fmt.Sprintf("ico.FromFile: failed to convert '%s'", path),
			img.Bounds().Dx(),
			img.Bounds().Dy(),
			pngData,
		)
	}
}

func FromImage(img image.Image) ([]byte, error) {
	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, img); err != nil {
		return nil, err
	}
	pngData := pngBuf.Bytes()

	return pngToIco(
		"ico.FromImage",
		img.Bounds().Dx(),
		img.Bounds().Dy(),
		pngData,
	)
}

func pngToIco(errPrefix string, width, height int, imageData []byte) ([]byte, error) {
	if width < 1 || width > 256 ||
		height < 1 || height > 256 {
		return nil, fmt.Errorf(
			"%s: illegal image size, width and height must be in range [1..256] but the given image has size %dx%d",
			errPrefix, width, height,
		)
	}

	var buf bytes.Buffer
	buf.Write([]byte{
		// ico header:
		0, 0, // reserved
		1, 0, // 1 for .ico, 2 for .cur
		1, 0, // number of images that follow

		// image header:
		byte(width),  // 0 means 256
		byte(height), // 0 means 256
		0,            // no color palette
		0,            // reserved
		1, 0,         // color planes
		32, 0, // bits per pixel
	})
	// image data size in bytes
	binary.Write(&buf, binary.LittleEndian, uint32(len(imageData)))
	// image data offset in ico file
	buf.Write([]byte{22, 0, 0, 0})
	// image data
	buf.Write(imageData)

	return buf.Bytes(), nil
}
