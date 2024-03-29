package frame

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

const (
	inputPath  = "input"
	outputPath = "output"
	DPI        = 300
)

var (
	A3Height = 0
	A3Width  = 0
	Padding  = 0
)

type imgType struct {
	image.Image
	name string
}

func Process(height, width, padding int, tChan chan int, okChan chan struct{}) {
	A3Height = calculatePixel(height)
	A3Width = calculatePixel(width)
	Padding = calculatePixel(padding)

	cpu := runtime.GOMAXPROCS(0)
	pool, _ := ants.NewPool(cpu)

	//exePath, _ := os.Executable()
	//exeDir := filepath.Dir(exePath)
	//inputPathAbs := filepath.Join(exeDir, inputPath)
	//outputPathAbs := filepath.Join(exeDir, outputPath)

	inputPathAbs := "input"
	outputPathAbs := "output"

	categories, err := os.ReadDir(inputPathAbs)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	totalCount := 0

	for _, category := range categories {
		categoryPath := filepath.Join(inputPathAbs, category.Name())
		if err != nil {
			panic(err)
		}

		images, err := os.ReadDir(categoryPath)
		if err != nil {
			panic(err)
		}

		totalCount += len(images)
		tChan <- totalCount

		go func(c os.DirEntry) {
			for _, image := range images {
				wg.Add(1)
				imageInputPath := filepath.Join(categoryPath, image.Name())
				imageOutputPath := filepath.Join(outputPathAbs, c.Name())

				pool.Submit(func() {
					processImage(image, imageInputPath, imageOutputPath)

					fmt.Println("Processed: ", image.Name())
					okChan <- struct{}{}

					wg.Done()
				})
			}
		}(category)
	}

	wg.Wait()

	fmt.Println("All images processed")
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

	return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
}

func placeImageToA3(img image.Image) image.Image {
	a3 := image.NewRGBA(image.Rect(0, 0, A3Width, A3Height))

	offsetX := (a3.Bounds().Dx() - img.Bounds().Dx()) / 2
	offsetY := (a3.Bounds().Dy() - img.Bounds().Dy()) / 2

	white := color.RGBA{255, 255, 255, 255}
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

func saveImage(img imgType, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			panic(err)
		}
	}

	out, err := os.Create(filepath.Join(path, img.name))
	if err != nil {
		panic(err)
	}
	defer out.Close()

	png.Encode(out, img.Image)
}

func processImage(file os.DirEntry, fullPath, outputPathAbs string) {
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
	}, outputPathAbs)
}

func calculatePixel(mm int) int {
	return int(float32(mm) * (float32(DPI) / float32(25.4)))
}
