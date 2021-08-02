package photopdf

import (
	"fmt"
	"github.com/nfnt/resize"
	"image"
	"image/jpeg"
	"math"
	"os"
)
import "path/filepath"
import "github.com/signintech/gopdf"


func Convert(source string) {
	fmt.Println("Reading files from (", source, ") and saving the result assets (", source,")")
	fmt.Println("-----------------------")
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4 })
	jpgs, _ := filepath.Glob(source + "/*.jpg")
	pngs, _ := filepath.Glob(source + "/*.png")
	jpegs, _ := filepath.Glob(source + "/*.jpeg")
	files := append(jpgs, append(jpegs, pngs...)...)
	for i := 0; i < len(files); i++ {
		fmt.Println(i+1, "adding ", files[i])
		x := float64(0)
		if x < 0 {
			continue
		}
		pdf.AddPage()
		goodSize(files[i])
		pdf.Image(files[i], 0, 0, nil)
	}
	fmt.Println("saving to ", source, " ...")
	pdf.WritePdf(fmt.Sprintf("%s/result.pdf", source))

	fmt.Println("-----------------------")
	fmt.Println("Done, have fun ;)")
}



const DefaultMaxWidth float64 = 595 * 1.6
const DefaultMaxHeight float64 = 842 * 1.6

// Рассчитываем размер изображения после масштабирования
func calculateRatioFit(sourceWidth, sourceHeight int) (int, int) {
	ratio := math.Min(DefaultMaxWidth/float64(sourceWidth), DefaultMaxHeight/float64(sourceHeight))
	return int(math.Ceil(float64(sourceWidth) * ratio)), int(math.Ceil(float64(sourceHeight) * ratio))
}

func goodSize(file string) {
	photo, _ := os.Open(file)
	defer photo.Close()

	img, _, err := image.Decode(photo)
	if err != nil {
		return
	}

	b := img.Bounds()
	width := b.Max.X
	height := b.Max.Y

	w, h := calculateRatioFit(width, height)

	fmt.Println("width = ", width, " height = ", height)
	fmt.Println("w = ", w, " h = ", h)

	// Вызов библиотеки изменения размера изображения
	m := resize.Resize(uint(w), uint(h), img, resize.Lanczos3)

	// файл для сохранения
	imgfile, _ := os.Create(file)
	defer imgfile.Close()

	// Сохраняем файл в формате jpeg
	err = jpeg.Encode(imgfile, m, nil)
	if err != nil {
		return
	}
}

