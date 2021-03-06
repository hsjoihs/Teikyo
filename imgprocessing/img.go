package imageprocessing

import (
	"image"
	"image/draw"
	"image/png"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hatobus/Teikyo/models"
	"github.com/nfnt/resize"
)

func GenTeikyo(fstream multipart.File, parts *models.Landmark, multi bool, n int, cnt int) error {

	goPath := os.Getenv("GOPATH")
	imgPath := filepath.Join(goPath, "src", "github.com", "hatobus", "Teikyo", "picture", "material")
	outputfile := filepath.Join(goPath, "src", "github.com", "hatobus", "Teikyo", "picture", "output", "output"+strconv.Itoa(n)+".png")

	t, _ := os.Open(filepath.Join(imgPath, "teikyo-tei.png"))
	defer t.Close()

	k, _ := os.Open(filepath.Join(imgPath, "teikyo-kyo.png"))
	defer k.Close()

	if _, err := fstream.Seek(0, 0); err != nil {
		return err
	}

	Tei, _, err := image.Decode(t)
	if err != nil {
		return err
	}

	Kyo, _, err := image.Decode(k)
	if err != nil {
		return err
	}

	dstimg, _, err := image.Decode(fstream)
	if err != nil {
		return err
	}

	if cnt > 0 && multi {
		t, err := os.Open(outputfile)
		if err != nil {
			return err
		}

		dstimg, _, err = image.Decode(t)
		if err != nil {
			return err
		}
	}

	// 大きさの微調整
	RightEyeSizeWidth := parts.EyeRight.BottomX - parts.EyeRight.TopX + 15
	RightEyeSizeHeight := parts.EyeRight.BottomY - parts.EyeRight.TopY + 15

	LeftEyeSizeWidth := parts.EyeLeft.BottomX - parts.EyeLeft.TopX + 15
	LeftEyeSizeHeight := parts.EyeLeft.BottomY - parts.EyeLeft.TopY + 15

	// 目の周りを覆うように大きさをリサイズ
	LTei := resize.Resize(
		uint(math.Abs(LeftEyeSizeWidth)),
		uint(math.Abs(LeftEyeSizeHeight)),
		Tei,
		resize.Lanczos3)

	RKyo := resize.Resize(
		uint(math.Abs(RightEyeSizeWidth)),
		uint(math.Abs(RightEyeSizeHeight)),
		Kyo,
		resize.Lanczos3)

	// References this
	// http://dempatow.hatenablog.com

	// 描画を開始する場所
	TeistartPoint := image.Point{int(parts.EyeLeft.TopX) - 10, int(parts.EyeLeft.TopY)}
	KyostartPoint := image.Point{int(parts.EyeRight.TopX) - 10, int(parts.EyeRight.TopY)}

	// 画像を入れる部分を作る
	TeiRectangle := image.Rectangle{TeistartPoint, TeistartPoint.Add(Tei.Bounds().Size())}
	KyoRectangle := image.Rectangle{KyostartPoint, KyostartPoint.Add(Kyo.Bounds().Size())}
	DstRectangle := image.Rectangle{image.Point{0, 0}, dstimg.Bounds().Size()}

	// 実際に描画、draw.Overで透過したまま重ねることができる
	rgba := image.NewRGBA(DstRectangle)
	draw.Draw(rgba, DstRectangle, dstimg, image.Pt(0, 0), draw.Src)
	draw.Draw(rgba, TeiRectangle, LTei, image.Pt(0, 0), draw.Over)
	draw.Draw(rgba, KyoRectangle, RKyo, image.Pt(0, 0), draw.Over)

	outfile, err := os.Create(outputfile)
	defer outfile.Close()
	if err != nil {
		return err
	}

	png.Encode(outfile, rgba)

	return nil
}
