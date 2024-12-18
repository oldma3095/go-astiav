package astiav

//#include <libavutil/imgutils.h>
//#include <libavutil/samplefmt.h>
//#include <stdlib.h>
//#include "macros.h"
import "C"
import (
	"errors"
	"fmt"
	"image"
	"strings"
	"unsafe"
)

type FrameData struct {
	f frameDataFramer
}

type frameDataFramer interface {
	bytes(align int) ([]byte, error)
	copyPlanes(ps []frameDataPlane) error
	height() int
	pixelFormat() PixelFormat
	planes(b []byte, align int) ([]frameDataPlane, error)
	width() int
}

type frameDataPlane struct {
	bytes    []byte
	linesize int
}

func newFrameData(f frameDataFramer) *FrameData {
	return &FrameData{f: f}
}

func (d *FrameData) Bytes(align int) ([]byte, error) {
	return d.f.bytes(align)
}

// It's the developer's responsibility to handle frame's writability
func (d *FrameData) SetBytes(b []byte, align int) error {
	// Get planes
	planes, err := d.f.planes(b, align)
	if err != nil {
		return fmt.Errorf("astiav: getting planes failed: %w", err)
	}

	// Copy planes
	if err := d.f.copyPlanes(planes); err != nil {
		return fmt.Errorf("astiav: copying planes failed: %w", err)
	}
	return nil
}

// Always returns non-premultiplied formats when dealing with alpha channels, however this might not
// always be accurate. In this case, use your own format in .ToImage()
func (d *FrameData) GuessImageFormat() (image.Image, error) {
	switch d.f.pixelFormat() {
	case PixelFormatGray8:
		return &image.Gray{}, nil
	case PixelFormatGray16Be:
		return &image.Gray16{}, nil
	case PixelFormatRgb0, PixelFormat0Rgb, PixelFormatRgb4, PixelFormatRgb8:
		return &image.RGBA{}, nil
	case PixelFormatRgba:
		return &image.NRGBA{}, nil
	case PixelFormatRgba64Be:
		return &image.NRGBA64{}, nil
	case PixelFormatYuva420P,
		PixelFormatYuva422P,
		PixelFormatYuva444P:
		return &image.NYCbCrA{}, nil
	case PixelFormatYuv410P,
		PixelFormatYuv411P, PixelFormatYuvj411P,
		PixelFormatYuv420P, PixelFormatYuvj420P,
		PixelFormatYuv422P, PixelFormatYuvj422P,
		PixelFormatYuv440P, PixelFormatYuvj440P,
		PixelFormatYuv444P, PixelFormatYuvj444P:
		return &image.YCbCr{}, nil
	}
	return nil, fmt.Errorf("astiav: pixel format %s not handled by Go", d.f.pixelFormat())
}

func (d *FrameData) imageYCbCrSubsampleRatio() image.YCbCrSubsampleRatio {
	name := d.f.pixelFormat().Name()
	for s, r := range map[string]image.YCbCrSubsampleRatio{
		"410": image.YCbCrSubsampleRatio410,
		"411": image.YCbCrSubsampleRatio411,
		"420": image.YCbCrSubsampleRatio420,
		"422": image.YCbCrSubsampleRatio422,
		"440": image.YCbCrSubsampleRatio440,
		"444": image.YCbCrSubsampleRatio444,
	} {
		if strings.Contains(name, s) {
			return r
		}
	}
	return image.YCbCrSubsampleRatio444
}

func (d *FrameData) toImagePix(pix *[]uint8, stride *int, rect *image.Rectangle, planes []frameDataPlane) {
	*pix = planes[0].bytes
	if v := planes[0].linesize; *stride != v {
		*stride = v
	}
	if w, h := d.f.width(), d.f.height(); rect.Dy() != w || rect.Dx() != h {
		*rect = image.Rect(0, 0, w, h)
	}
}

func (d *FrameData) toImageYCbCr(y, cb, cr *[]uint8, yStride, cStride *int, subsampleRatio *image.YCbCrSubsampleRatio, rect *image.Rectangle, planes []frameDataPlane) {
	*y = planes[0].bytes
	*cb = planes[1].bytes
	*cr = planes[2].bytes
	if v := planes[0].linesize; *yStride != v {
		*yStride = v
	}
	if v := planes[1].linesize; *cStride != v {
		*cStride = v
	}
	if v := d.imageYCbCrSubsampleRatio(); *subsampleRatio != v {
		*subsampleRatio = v
	}
	if w, h := d.f.width(), d.f.height(); rect.Dy() != w || rect.Dx() != h {
		*rect = image.Rect(0, 0, w, h)
	}
}

func (d *FrameData) toImageYCbCrA(y, cb, cr, a *[]uint8, yStride, cStride, aStride *int, subsampleRatio *image.YCbCrSubsampleRatio, rect *image.Rectangle, planes []frameDataPlane) {
	d.toImageYCbCr(y, cb, cr, yStride, cStride, subsampleRatio, rect, planes)
	*a = planes[3].bytes
	if v := planes[3].linesize; *aStride != v {
		*aStride = v
	}
}

func (d *FrameData) ToImage(dst image.Image) error {
	// Get bytes
	// Using bytesFromC on f.c.data caused random segfaults
	const align = 1
	b, err := d.f.bytes(align)
	if err != nil {
		return fmt.Errorf("astiav: getting bytes failed: %w", err)
	}

	// Get planes
	planes, err := d.f.planes(b, align)
	if err != nil {
		return fmt.Errorf("astiav: getting planes failed: %w", err)
	}

	// Update image
	switch v := dst.(type) {
	case *image.Alpha:
		d.toImagePix(&v.Pix, &v.Stride, &v.Rect, planes)
	case *image.Alpha16:
		d.toImagePix(&v.Pix, &v.Stride, &v.Rect, planes)
	case *image.CMYK:
		d.toImagePix(&v.Pix, &v.Stride, &v.Rect, planes)
	case *image.Gray:
		d.toImagePix(&v.Pix, &v.Stride, &v.Rect, planes)
	case *image.Gray16:
		d.toImagePix(&v.Pix, &v.Stride, &v.Rect, planes)
	case *image.NRGBA:
		d.toImagePix(&v.Pix, &v.Stride, &v.Rect, planes)
	case *image.NRGBA64:
		d.toImagePix(&v.Pix, &v.Stride, &v.Rect, planes)
	case *image.NYCbCrA:
		d.toImageYCbCrA(&v.Y, &v.Cb, &v.Cr, &v.A, &v.YStride, &v.CStride, &v.AStride, &v.SubsampleRatio, &v.Rect, planes)
	case *image.RGBA:
		d.toImagePix(&v.Pix, &v.Stride, &v.Rect, planes)
	case *image.RGBA64:
		d.toImagePix(&v.Pix, &v.Stride, &v.Rect, planes)
	case *image.YCbCr:
		d.toImageYCbCr(&v.Y, &v.Cb, &v.Cr, &v.YStride, &v.CStride, &v.SubsampleRatio, &v.Rect, planes)
	default:
		return errors.New("astiav: image format is not handled")
	}
	return nil
}

func (d *FrameData) fromImagePix(pix []uint8, stride int) error {
	// Copy planes
	if err := d.f.copyPlanes([]frameDataPlane{{bytes: pix, linesize: stride}}); err != nil {
		return fmt.Errorf("astiav: copying planes failed: %w", err)
	}
	return nil
}

func (d *FrameData) fromImageYCbCr(y, cb, cr []uint8, yStride, cStride int) error {
	// Copy planes
	if err := d.f.copyPlanes([]frameDataPlane{
		{bytes: y, linesize: yStride},
		{bytes: cb, linesize: cStride},
		{bytes: cr, linesize: cStride},
	}); err != nil {
		return fmt.Errorf("astiav: copying planes failed: %w", err)
	}
	return nil
}

func (d *FrameData) fromImageYCbCrA(y, cb, cr, a []uint8, yStride, cStride, aStride int) error {
	// Copy planes
	if err := d.f.copyPlanes([]frameDataPlane{
		{bytes: y, linesize: yStride},
		{bytes: cb, linesize: cStride},
		{bytes: cr, linesize: cStride},
		{bytes: a, linesize: aStride},
	}); err != nil {
		return fmt.Errorf("astiav: copying planes failed: %w", err)
	}
	return nil
}

// It's the developer's responsibility to handle frame's writability
func (d *FrameData) FromImage(src image.Image) error {
	// Copy planes
	switch v := src.(type) {
	case *image.Alpha:
		return d.fromImagePix(v.Pix, v.Stride)
	case *image.Alpha16:
		return d.fromImagePix(v.Pix, v.Stride)
	case *image.CMYK:
		return d.fromImagePix(v.Pix, v.Stride)
	case *image.Gray:
		return d.fromImagePix(v.Pix, v.Stride)
	case *image.Gray16:
		return d.fromImagePix(v.Pix, v.Stride)
	case *image.NRGBA:
		return d.fromImagePix(v.Pix, v.Stride)
	case *image.NRGBA64:
		return d.fromImagePix(v.Pix, v.Stride)
	case *image.NYCbCrA:
		return d.fromImageYCbCrA(v.Y, v.Cb, v.Cr, v.A, v.YStride, v.CStride, v.AStride)
	case *image.RGBA:
		return d.fromImagePix(v.Pix, v.Stride)
	case *image.RGBA64:
		return d.fromImagePix(v.Pix, v.Stride)
	case *image.YCbCr:
		return d.fromImageYCbCr(v.Y, v.Cb, v.Cr, v.YStride, v.CStride)
	}
	return errors.New("astiav: image format is not handled")
}

var _ frameDataFramer = (*frameDataFrame)(nil)

type frameDataFrame struct {
	f *Frame
}

func newFrameDataFrame(f *Frame) *frameDataFrame {
	return &frameDataFrame{f: f}
}

func (f *frameDataFrame) bytes(align int) ([]byte, error) {
	// Get funcs
	var bufferSizeFunc func(int) (int, error)
	var copyToBufferFunc func([]byte, int) (int, error)
	switch f.mediaType() {
	case MediaTypeAudio:
		bufferSizeFunc = f.f.SamplesBufferSize
		copyToBufferFunc = f.f.SamplesCopyToBuffer
	case MediaTypeVideo:
		bufferSizeFunc = f.f.ImageBufferSize
		copyToBufferFunc = f.f.ImageCopyToBuffer
	default:
		return nil, errors.New("astiav: media type not implemented")
	}

	// Get buffer size
	s, err := bufferSizeFunc(align)
	if err != nil {
		return nil, fmt.Errorf("astiav: getting buffer size failed: %w", err)
	}

	// Invalid buffer size
	if s == 0 {
		return nil, errors.New("astiav: invalid buffer size")
	}

	// Create buffer
	b := make([]byte, s)

	// Copy to buffer
	if _, err = copyToBufferFunc(b, align); err != nil {
		return nil, fmt.Errorf("astiav: copying to buffer failed: %w", err)
	}
	return b, nil
}

func (f *frameDataFrame) copyPlanes(ps []frameDataPlane) error {
	// Check writability
	if !f.f.IsWritable() {
		return errors.New("astiav: frame is not writable")
	}

	// Prepare data
	var cdata [8]*C.uint8_t
	var clinesizes [8]C.int
	for i, p := range ps {
		// Convert data
		if len(p.bytes) > 0 {
			cdata[i] = (*C.uint8_t)(C.CBytes(p.bytes))
			defer C.free(unsafe.Pointer(cdata[i]))
		}

		// Convert linesize
		clinesizes[i] = C.int(p.linesize)
	}

	// Copy data
	switch f.mediaType() {
	case MediaTypeAudio:
		C.av_samples_copy(&f.f.c.data[0], &cdata[0], 0, 0, f.f.c.nb_samples, f.f.c.ch_layout.nb_channels, (C.enum_AVSampleFormat)(f.f.c.format))
	case MediaTypeVideo:
		C.av_image_copy(&f.f.c.data[0], &f.f.c.linesize[0], &cdata[0], &clinesizes[0], (C.enum_AVPixelFormat)(f.f.c.format), f.f.c.width, f.f.c.height)
	default:
		return errors.New("astiav: media type not implemented")
	}
	return nil
}

func (f *frameDataFrame) height() int {
	return f.f.Height()
}

func (f *frameDataFrame) mediaType() MediaType {
	switch {
	// Audio
	case f.f.NbSamples() > 0:
		return MediaTypeAudio
	// Video
	case f.f.Height() > 0 && f.f.Width() > 0:
		return MediaTypeVideo
	default:
		return MediaTypeUnknown
	}
}

func (f *frameDataFrame) pixelFormat() PixelFormat {
	return f.f.PixelFormat()
}

func (f *frameDataFrame) planes(b []byte, align int) ([]frameDataPlane, error) {
	// Get line and plane sizes
	var linesizes [8]int
	var planeSizes [8]int
	switch f.mediaType() {
	case MediaTypeAudio:
		// Get buffer size
		var cLinesize C.int
		cBufferSize := C.av_samples_get_buffer_size(&cLinesize, f.f.c.ch_layout.nb_channels, f.f.c.nb_samples, (C.enum_AVSampleFormat)(f.f.c.format), C.int(align))
		if err := newError(cBufferSize); err != nil {
			return nil, fmt.Errorf("astiav: getting buffer size failed: %w", err)
		}

		// Update line and plane sizes
		for i := 0; i < int(cBufferSize/cLinesize); i++ {
			linesizes[i] = int(cLinesize)
			planeSizes[i] = int(cLinesize)
		}
	case MediaTypeVideo:
		// Below is mostly inspired by https://ffmpeg.org/doxygen/7.0/group__lavu__picture.html#ga24a67963c3ae0054a2a4bab35930e694

		// Get linesize
		var cLinesizes [8]C.int
		if err := newError(C.av_image_fill_linesizes(&cLinesizes[0], (C.enum_AVPixelFormat)(f.f.c.format), f.f.c.width)); err != nil {
			return nil, fmt.Errorf("astiav: getting linesize failed: %w", err)
		}

		// Align linesize
		var cAlignedLinesizes [8]C.ptrdiff_t
		for i := 0; i < 4; i++ {
			cAlignedLinesizes[i] = C.astiavFFAlign(cLinesizes[i], C.int(align))
		}

		// Get plane sizes
		var cPlaneSizes [8]C.size_t
		if err := newError(C.av_image_fill_plane_sizes(&cPlaneSizes[0], (C.enum_AVPixelFormat)(f.f.c.format), f.f.c.height, &cAlignedLinesizes[0])); err != nil {
			return nil, fmt.Errorf("astiav: getting plane sizes failed: %w", err)
		}

		// Update line and plane sizes
		for i := range cPlaneSizes {
			linesizes[i] = int(cAlignedLinesizes[i])
			planeSizes[i] = int(cPlaneSizes[i])
		}
	default:
		return nil, errors.New("astiav: media type not implemented")
	}

	// Loop through plane sizes
	var ps []frameDataPlane
	start := 0
	for i := range planeSizes {
		// Get end
		end := start + planeSizes[i]
		if len(b) < end {
			return nil, fmt.Errorf("astiav: buffer length %d is invalid for [%d:%d]", len(b), start, end)
		}

		// Append plane
		ps = append(ps, frameDataPlane{
			bytes:    b[start:end],
			linesize: linesizes[i],
		})

		// Update start
		start = end
	}
	return ps, nil
}

func (f *frameDataFrame) width() int {
	return f.f.Width()
}
