package core

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
)

// GenerateCollage make big image from the input images.
// Height and Width is the amount of image tiles.
// If height = 2 and width = 3 then collage image will be 1280x1920px
func GenerateCollage(images []image.Image, height int, width int) image.Image {
	if len(images) != height*width {
		log.Fatalf("Need %d pixs, founded %d", height*width, len(images))
	}

	canvasImage := image.NewRGBA(image.Rect(0, 0, 640*width, 640*height))
	fmt.Printf("Found %d pics", len(images))
	for index, img := range images {
		var x, y int
		if width > height {
			x = 640 * (index % width)
			y = 640 * (index / width)
		} else {
			x = 640 * (index / height)
			y = 640 * (index % height)
		}
		log.Printf("%d x %d", x, y)
		img = Resize(CropToSquare(img), 640)
		draw.Draw(canvasImage, img.Bounds().Add(image.Pt(x, y)), img, image.ZP, draw.Src)
	}

	return image.Image(canvasImage)
}

//Resize image to sizes
func Resize(img image.Image, width uint) image.Image {
	if img.Bounds().Size().X == int(width) {
		return img
	}
	return resize.Resize(width, 0, img, resize.Lanczos3)
}

//CropToSquare crop image so it be square
func CropToSquare(img image.Image) image.Image {
	width := img.Bounds().Size().X
	height := img.Bounds().Size().Y
	if width == height {
		return img
	}
	var x, y, size int
	if width > height {
		x = (width - height) / 2
		y = 0
		size = height
	} else {
		x = 0
		y = (height - width) / 2
		size = width
	}
	canvas := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(canvas, canvas.Bounds(), img, image.Point{x, y}, draw.Src)
	return canvas
}

// IsSquare checks if image is real square or with white/black frame.
// TODO: it's quite stupid algorythm, but toss away 95% of bad pics
func IsSquare(image image.Image) bool {
	bounds := image.Bounds()
	width := bounds.Size().X
	height := bounds.Size().Y
	alignmentX := 1
	alignmentY := 1

	// check left column
	for x := bounds.Min.X; x < bounds.Min.X+6; x++ {
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			r, g, b, _ := image.At(x, y).RGBA()
			rgb := (r + g + b) / 255
			if isFrameColor(rgb) {
				alignmentY++
				if alignmentY > height*5 {
					return false
				}
			} else {
				alignmentY = 1
			}
		}
	}

	// check upper row
	for y := bounds.Min.Y; y < bounds.Min.Y+6; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := image.At(x, y).RGBA()
			rgb := (r + g + b) / 255
			if isFrameColor(rgb) {
				alignmentX++
				if alignmentX > width*5 {
					return false
				}
			} else {
				alignmentX = 1
			}
		}
	}

	return true
}

func isFrameColor(rgb uint32) bool {
	return rgb < 15 || rgb > 240
}
