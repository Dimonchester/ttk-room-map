package space

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"

	"github.com/Dimonchester/ttk-room-map/backend/internal/logger"
)

type MediaProcessor struct {
	uploadDir string
}

func NewMediaProcessor(uploadDir string) (*MediaProcessor, error) {
	for _, sub := range []string{"plans", "panoramas"} {
		if err := os.MkdirAll(filepath.Join(uploadDir, sub), 0755); err != nil {
			return nil, fmt.Errorf("create upload subdir %s: %w", sub, err)
		}
	}
	return &MediaProcessor{uploadDir: uploadDir}, nil
}

func (mp *MediaProcessor) ProcessPlanImage(mapID, filename string, r io.Reader) (*PlanImage, error) {
	logger.Info(logger.CatMedia, "обработка 2D-плана для карты map_id=%s file=%s", mapID, filename)

	img, format, err := image.Decode(r)
	if err != nil {
		logger.Error(logger.CatMedia, "ошибка декодирования изображения 2D-плана map_id=%s: %v", mapID, err)
		return nil, fmt.Errorf("decode plan image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	outName := fmt.Sprintf("plan_%s.jpg", mapID)
	outPath := filepath.Join(mp.uploadDir, "plans", outName)

	outFile, err := os.Create(outPath)
	if err != nil {
		logger.Error(logger.CatMedia, "ошибка создания файла 2D-плана %s: %v", outPath, err)
		return nil, err
	}
	defer outFile.Close()

	if err := jpeg.Encode(outFile, img, &jpeg.Options{Quality: 90}); err != nil {
		logger.Error(logger.CatMedia, "ошибка сохранения JPEG 2D-плана %s: %v", outPath, err)
		return nil, err
	}

	logger.Info(logger.CatMedia, "2D-план сохранен map_id=%s format=%s size=%dx%d", mapID, format, width, height)
	return &PlanImage{
		ImageURL: "/uploads/plans/" + outName,
		Width:    width,
		Height:   height,
	}, nil
}

func (mp *MediaProcessor) ProcessUploadedPanorama(panoramaID string, src image.Image) (string, string, int, int, error) {
	logger.Info(logger.CatMedia, "генерация полной 3D-панорамы и превью для panorama_id=%s", panoramaID)

	fullWidth := 2048
	fullHeight := 1024
	fullImg := resizeBilinear(src, fullWidth, fullHeight)
	previewImg := resizeBilinear(fullImg, 512, 256)

	fullURL, prevURL, err := mp.savePanoramaFiles(panoramaID, fullImg, previewImg)
	if err != nil {
		return "", "", 0, 0, err
	}

	logger.Info(logger.CatMedia, "готовая панорама обработана panorama_id=%s (%dx%d)", panoramaID, fullWidth, fullHeight)
	return fullURL, prevURL, fullWidth, fullHeight, nil
}

func (mp *MediaProcessor) StitchFramesToPanorama(panoramaID string, frames []image.Image) (string, string, int, int, error) {
	logger.Info(logger.CatMedia, "старт сшивки 3D-панорамы panorama_id=%s из %d кадров", panoramaID, len(frames))

	if len(frames) == 0 {
		return "", "", 0, 0, fmt.Errorf("no frames provided")
	}

	const (
		panoW = 2048
		panoH = 1024
	)

	canvas := image.NewRGBA(image.Rect(0, 0, panoW, panoH))
	n := len(frames)
	overlap := panoW / (n * 6)
	if overlap < 16 {
		overlap = 16
	}

	segmentW := panoW / n

	for i, frame := range frames {
		targetW := segmentW + overlap
		resized := resizeBilinear(frame, targetW, panoH)
		startX := i * segmentW

		for y := 0; y < panoH; y++ {
			for x := 0; x < targetW; x++ {
				destX := (startX + x) % panoW
				srcCol := resized.RGBAAt(x, y)

				alpha := 1.0
				if i > 0 && x < overlap {
					alpha = float64(x) / float64(overlap)
				}

				if alpha >= 0.999 {
					canvas.SetRGBA(destX, y, srcCol)
				} else {
					existing := canvas.RGBAAt(destX, y)
					blended := color.RGBA{
						R: uint8(float64(existing.R)*(1-alpha) + float64(srcCol.R)*alpha),
						G: uint8(float64(existing.G)*(1-alpha) + float64(srcCol.G)*alpha),
						B: uint8(float64(existing.B)*(1-alpha) + float64(srcCol.B)*alpha),
						A: 255,
					}
					canvas.SetRGBA(destX, y, blended)
				}
			}
		}
	}

	previewImg := resizeBilinear(canvas, 512, 256)
	fullURL, prevURL, err := mp.savePanoramaFiles(panoramaID, canvas, previewImg)
	if err != nil {
		return "", "", 0, 0, err
	}

	logger.Info(logger.CatMedia, "сшивка 3D-панорамы завершена успешно panorama_id=%s", panoramaID)
	return fullURL, prevURL, panoW, panoH, nil
}

func (mp *MediaProcessor) savePanoramaFiles(panoramaID string, fullImg, previewImg image.Image) (string, string, error) {
	fullName := fmt.Sprintf("pano_%s_full.jpg", panoramaID)
	prevName := fmt.Sprintf("pano_%s_preview.jpg", panoramaID)

	fullPath := filepath.Join(mp.uploadDir, "panoramas", fullName)
	prevPath := filepath.Join(mp.uploadDir, "panoramas", prevName)

	fFull, err := os.Create(fullPath)
	if err != nil {
		logger.Error(logger.CatMedia, "ошибка записи файла панорамы %s: %v", fullPath, err)
		return "", "", err
	}
	defer fFull.Close()

	if err := jpeg.Encode(fFull, fullImg, &jpeg.Options{Quality: 88}); err != nil {
		logger.Error(logger.CatMedia, "ошибка кодирования JPEG панорамы %s: %v", fullPath, err)
		return "", "", err
	}

	fPrev, err := os.Create(prevPath)
	if err != nil {
		logger.Error(logger.CatMedia, "ошибка записи превью панорамы %s: %v", prevPath, err)
		return "", "", err
	}
	defer fPrev.Close()

	if err := jpeg.Encode(fPrev, previewImg, &jpeg.Options{Quality: 75}); err != nil {
		logger.Error(logger.CatMedia, "ошибка кодирования превью панорамы %s: %v", prevPath, err)
		return "", "", err
	}

	return "/uploads/panoramas/" + fullName, "/uploads/panoramas/" + prevName, nil
}

func resizeBilinear(src image.Image, dstW, dstH int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, dstW, dstH))
	b := src.Bounds()
	srcW := b.Dx()
	srcH := b.Dy()
	if srcW == 0 || srcH == 0 {
		draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)
		return dst
	}

	for y := 0; y < dstH; y++ {
		sy := b.Min.Y + (y*srcH)/dstH
		for x := 0; x < dstW; x++ {
			sx := b.Min.X + (x*srcW)/dstW
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}
