package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/panjf2000/ants/v2"
)

// Pixels=Millimeters×(DPI/25.4)
const (
	inputPath  = "input"
	outputPath = "output"
	A3Width    = 3508
	A3Height   = 4961
	Padding    = 364.5
)

type imgType struct {
	image.Image
	name string
}

func main() {

	cpu := runtime.GOMAXPROCS(0)
	fmt.Printf("cpu: %v\n", cpu)

	pool, _ := ants.NewPool(cpu)
	inputPathAbs, err := filepath.Abs(inputPath)
	if err != nil {
		panic(err)
	}

	files, err := os.ReadDir(inputPathAbs)
	if err != nil {
		fmt.Print("asd")
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(files))

	for _, file := range files {
		pool.Submit(func() {
			fullPath := filepath.Join(inputPath, file.Name())
			fullPath, err = filepath.Abs(fullPath)
			if err != nil {
				panic(err)
			}

			imgFile, err := os.Open(fullPath)
			if err != nil {
				panic(err)
			}

			img, _, err := image.Decode(imgFile)
			if err != nil {
				panic(err)
			}

			resizedImage := calculateDimensions(img)
			placedImage := placeImageToA3(resizedImage)
			saveImage(imgType{
				Image: placedImage,
				name:  file.Name(),
			})

			wg.Done()
		})

	}

	wg.Wait()
}

func calculateDimensions(img image.Image) image.Image {
	originalWidth := img.Bounds().Dx()
	originalHeight := img.Bounds().Dy()
	isHorizontal := originalWidth > originalHeight

	var scale float32
	var newWidth, newHeight int

	if isHorizontal {
		img = imaging.Rotate90(img)
		originalWidth = img.Bounds().Dx()
		originalHeight = img.Bounds().Dy()
	}

	scale = float32(A3Height-2*Padding) / float32(originalHeight)
	newWidth = int(float32(originalWidth) * scale)
	newHeight = A3Height - 2*Padding

	for newWidth > A3Width-2*Padding {
		scale = float32(A3Width-2*Padding) / float32(newWidth)
		newHeight = int(float32(newHeight) * scale)
		newWidth = A3Width - 2*Padding
	}

	fmt.Printf("newHeight: %v\n", newHeight)
	fmt.Printf("newWidth: %v\n", newWidth)

	return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
}

func placeImageToA3(img image.Image) image.Image {
	a3 := image.NewRGBA(image.Rect(0, 0, A3Width, A3Height))

	offsetX := (a3.Bounds().Dx() - img.Bounds().Dx()) / 2
	offsetY := (a3.Bounds().Dy() - img.Bounds().Dy()) / 2

	white := color.RGBA{200, 200, 200, 255}
	for y := 0; y < a3.Bounds().Dy(); y++ {
		for x := 0; x < a3.Bounds().Dx(); x++ {
			a3.Set(x, y, white)
		}
	}

	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			a3.Set(x+offsetX, y+offsetY, img.At(x, y))
		}
	}

	return a3
}

func saveImage(img imgType) {
	out, err := os.Create(filepath.Join(outputPath, img.name))
	if err != nil {
		panic(err)
	}
	defer out.Close()

	png.Encode(out, img.Image)
}
