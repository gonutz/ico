package ico

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	_ "github.com/gonutz/bmp"
	_ "image/gif"
	_ "image/jpeg"
)

func FromImage(img image.Image) ([]byte, error) {
	data, err := imageToIco(img)
	if err != nil {
		return nil, fmt.Errorf("ico.FromImage: %w", err)
	}
	return data, nil
}

func FromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf(
			"ico.FromFile: failed to read file '%s': %w",
			path, err,
		)
	}

	if is32BitPNG(data) {
		return from32bitPngFile(path, data)
	} else {
		return fromImageFile(path, data)
	}
}

func from32bitPngFile(path string, data []byte) ([]byte, error) {
	// Use PNG file data as is without re-encoding it. This special case is
	// handled to potentially make the icon file smaller. The PNG file might
	// have been optimized, e.g. with ZopfliPNG, and we do not want to undo
	// that optimization by re-encoding a valid PNG file.
	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf(
			"ico.FromFile: failed to decode PNG header of file '%s': %w",
			path, err,
		)
	}

	icon, err := pngToIco(
		config.Width,
		config.Height,
		data,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"ico.FromFile: failed to decode PNG file '%s': %w",
			path, err,
		)
	}
	return icon, nil
}

func fromImageFile(path string, data []byte) ([]byte, error) {
	// Decode the image, which we do not know the format of.
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf(
			"ico.FromFile: failed to decode image file '%s': %w",
			path, err,
		)
	}

	icon, err := imageToIco(img)
	if err != nil {
		return nil, fmt.Errorf("ico.FromFile: failed to convert '%s': %w", path, err)
	}
	return icon, nil
}

func is32BitPNG(data []byte) bool {
	// Make sure this is a valid PNG file by decoding it and looking whether we
	// get an error.
	_, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return false
	}

	// Every PNG file starts with this magic header, directly followed by the
	// IHRD chunk, which contains the information we need.
	header := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0, 0, 0, 13, // IHDR chunk length
		'I', 'H', 'D', 'R', // IHDR chunk ID
	}
	if !bytes.HasPrefix(data, header) {
		return false
	}

	// The IHDR starts with 4 bytes width, 4 bytes height, then our desired
	// information: bit depth and color type, one byte each.
	if len(data) < len(header)+4+4+1+1 {
		return false
	}
	bitDepth := data[len(header)+4+4]
	colorType := data[len(header)+4+4+1]

	// We want 8 bits per channel and truecolor with alpha, meaning red, green,
	// blue and alpha channels.
	const truecolorWithAlpha = 6
	return bitDepth == 8 && colorType == truecolorWithAlpha
}

func imageToIco(img image.Image) ([]byte, error) {
	// Convert the image to a format that can hold RGBA information.
	b := img.Bounds()
	nrgba := image.NewNRGBA(b)
	draw.Draw(nrgba, b, img, b.Min, draw.Src)

	// Go's PNG encoder stores images that do not have any transparency as color
	// type RGB (ID 2 in the PNG's IHRD chunk). We need type RGBA (ID 6 in the
	// PNG's IHRD chunk) or the icon file will not be shown properyl. To force
	// this, we need at least one pixel in the image that is not fully opaque
	// (i.e. a pixel with alpha != 255).
	fullyOpaque := true
	end := nrgba.PixOffset(b.Max.X-1, b.Max.Y-1) + 4
	for a := 3; a < end; a += 4 {
		if nrgba.Pix[a] < 255 {
			fullyOpaque = false
			break
		}
	}
	if fullyOpaque {
		// In case the image is fully opaque, we arbitrarily make the
		// bottom-right pixel have minimal transparency, almost fully opaque.
		// This will force the PNG encoder to use RGBA colors.
		nrgba.Pix[end-1] = 254
	}

	var pngBuf bytes.Buffer
	if err := png.Encode(&pngBuf, nrgba); err != nil {
		return nil, err
	}
	pngData := pngBuf.Bytes()

	return pngToIco(b.Dx(), b.Dy(), pngData)
}

func pngToIco(width, height int, imageData []byte) ([]byte, error) {
	if width < 1 || width > 256 ||
		height < 1 || height > 256 {
		return nil, fmt.Errorf(
			"illegal image size, width and height must be in "+
				"range [1..256] but the given image has size %dx%d",
			width, height,
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
