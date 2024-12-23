package astiav

//#include "channel_layout.h"
import "C"
import (
	"unsafe"
)

// https://ffmpeg.org/doxygen/7.0/group__lavu__audio__channels.html#ga855bb7dede67971e95bd09d8fcca7293
var (
	ChannelLayoutMono              = newChannelLayoutFromC(C.astiavChannelLayoutMono)
	ChannelLayoutStereo            = newChannelLayoutFromC(C.astiavChannelLayoutStereo)
	ChannelLayout2Point1           = newChannelLayoutFromC(C.astiavChannelLayout2Point1)
	ChannelLayout21                = newChannelLayoutFromC(C.astiavChannelLayout21)
	ChannelLayoutSurround          = newChannelLayoutFromC(C.astiavChannelLayoutSurround)
	ChannelLayout3Point1           = newChannelLayoutFromC(C.astiavChannelLayout3Point1)
	ChannelLayout4Point0           = newChannelLayoutFromC(C.astiavChannelLayout4Point0)
	ChannelLayout4Point1           = newChannelLayoutFromC(C.astiavChannelLayout4Point1)
	ChannelLayout22                = newChannelLayoutFromC(C.astiavChannelLayout22)
	ChannelLayoutQuad              = newChannelLayoutFromC(C.astiavChannelLayoutQuad)
	ChannelLayout5Point0           = newChannelLayoutFromC(C.astiavChannelLayout5Point0)
	ChannelLayout5Point1           = newChannelLayoutFromC(C.astiavChannelLayout5Point1)
	ChannelLayout5Point0Back       = newChannelLayoutFromC(C.astiavChannelLayout5Point0Back)
	ChannelLayout5Point1Back       = newChannelLayoutFromC(C.astiavChannelLayout5Point1Back)
	ChannelLayout6Point0           = newChannelLayoutFromC(C.astiavChannelLayout6Point0)
	ChannelLayout6Point0Front      = newChannelLayoutFromC(C.astiavChannelLayout6Point0Front)
	ChannelLayoutHexagonal         = newChannelLayoutFromC(C.astiavChannelLayoutHexagonal)
	ChannelLayout3Point1Point2     = newChannelLayoutFromC(C.astiavChannelLayout3Point1Point2)
	ChannelLayout6Point1           = newChannelLayoutFromC(C.astiavChannelLayout6Point1)
	ChannelLayout6Point1Back       = newChannelLayoutFromC(C.astiavChannelLayout6Point1Back)
	ChannelLayout6Point1Front      = newChannelLayoutFromC(C.astiavChannelLayout6Point1Front)
	ChannelLayout7Point0           = newChannelLayoutFromC(C.astiavChannelLayout7Point0)
	ChannelLayout7Point0Front      = newChannelLayoutFromC(C.astiavChannelLayout7Point0Front)
	ChannelLayout7Point1           = newChannelLayoutFromC(C.astiavChannelLayout7Point1)
	ChannelLayout7Point1Wide       = newChannelLayoutFromC(C.astiavChannelLayout7Point1Wide)
	ChannelLayout7Point1WideBack   = newChannelLayoutFromC(C.astiavChannelLayout7Point1WideBack)
	ChannelLayout5Point1Point2Back = newChannelLayoutFromC(C.astiavChannelLayout5Point1Point2Back)
	ChannelLayoutOctagonal         = newChannelLayoutFromC(C.astiavChannelLayoutOctagonal)
	ChannelLayoutCube              = newChannelLayoutFromC(C.astiavChannelLayoutCube)
	ChannelLayout5Point1Point4Back = newChannelLayoutFromC(C.astiavChannelLayout5Point1Point4Back)
	ChannelLayout7Point1Point2     = newChannelLayoutFromC(C.astiavChannelLayout7Point1Point2)
	ChannelLayout7Point1Point4Back = newChannelLayoutFromC(C.astiavChannelLayout7Point1Point4Back)
	ChannelLayoutHexadecagonal     = newChannelLayoutFromC(C.astiavChannelLayoutHexadecagonal)
	ChannelLayoutStereoDownmix     = newChannelLayoutFromC(C.astiavChannelLayoutStereoDownmix)
	ChannelLayout22Point2          = newChannelLayoutFromC(C.astiavChannelLayout22Point2)
	ChannelLayout7Point1TopBack    = newChannelLayoutFromC(C.astiavChannelLayout7Point1TopBack)
)

// https://ffmpeg.org/doxygen/7.0/structAVChannelLayout.html
type ChannelLayout struct {
	c *C.AVChannelLayout
}

func newChannelLayoutFromC(c *C.AVChannelLayout) ChannelLayout {
	return ChannelLayout{c: c}
}

// https://ffmpeg.org/doxygen/7.0/structAVChannelLayout.html#adfd3f460a8ea1575baa32852d9248d3c
func (l ChannelLayout) Channels() int {
	if l.c == nil {
		return 0
	}
	return int(l.c.nb_channels)
}

func (l ChannelLayout) String() string {
	b := make([]byte, 1024)
	n, err := l.Describe(b)
	if err != nil {
		return ""
	}
	return string(b[:n])
}

// https://ffmpeg.org/doxygen/7.0/group__lavu__audio__channels.html#gacc7d7d1a280248aafb8f9196c9d4e24f
func (l ChannelLayout) Describe(b []byte) (int, error) {
	if l.c == nil {
		return 0, nil
	}
	ret := C.av_channel_layout_describe(l.c, (*C.char)(unsafe.Pointer(&b[0])), C.size_t(len(b)))
	if err := newError(ret); err != nil {
		return 0, err
	}
	if ret > 0 && b[ret-1] == '\x00' {
		ret -= 1
	}
	return int(ret), nil
}

// https://ffmpeg.org/doxygen/7.0/group__lavu__audio__channels.html#gad15a6bf80ee8551ee4a4789d970ccbea
func (l ChannelLayout) Valid() bool {
	if l.c == nil {
		return false
	}
	return C.av_channel_layout_check(l.c) > 0
}

// https://ffmpeg.org/doxygen/7.0/group__lavu__audio__channels.html#ga5da99475fc07b778522974a2e0a1f58c
func (l ChannelLayout) Compare(l2 ChannelLayout) (equal bool, err error) {
	if l.c == nil || l2.c == nil {
		return l.c == nil && l2.c == nil, nil
	}
	ret := C.av_channel_layout_compare(l.c, l2.c)
	if err := newError(ret); err != nil {
		return false, err
	}
	return ret == 0, nil
}

func (l ChannelLayout) Equal(l2 ChannelLayout) bool {
	v, _ := l.Compare(l2)
	return v
}

// https://ffmpeg.org/doxygen/7.0/group__lavu__audio__channels.html#gad36be43b2a1b14b66492b8025b82f886
func (l ChannelLayout) copy(dst *C.AVChannelLayout) error {
	return newError(C.av_channel_layout_copy(dst, l.c))
}

func (l ChannelLayout) clone() (ChannelLayout, error) {
	var cl C.AVChannelLayout
	err := l.copy(&cl)
	dst := newChannelLayoutFromC(&cl)
	return dst, err
}
