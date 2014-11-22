package core

import (
	"fmt"
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

	canvas_image := image.NewRGBA(image.Rect(0, 0, 640*width, 640*height))
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
		draw.Draw(canvas_image, img.Bounds().Add(image.Pt(x, y)), img, image.ZP, draw.Src)
	}

	return image.Image(canvas_image)
}
