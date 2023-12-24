package libvips

/*
#cgo pkg-config: glib-2.0 gobject-2.0
#cgo LDFLAGS: -lvips
#include "vips/vips.h"
#include <stdio.h>

void libvips_error() {
	printf("libvips: error: %s\n", vips_error_buffer());
	vips_error_clear();
}

int libvips_init() {
	if (VIPS_INIT("")) {
		libvips_error();
		return -1;
	}
	return 0;
}

const char* libvips_version() {
	return vips_version_string();
}

void libvips_shutdown() {
	vips_shutdown();
}

int libvips_encode(char* buf, int len, void** out_buf, size_t* out_size, int outType, int animated) {
	VipsImage* img;
	if (animated) {
		img = vips_image_new_from_buffer(buf, len, "", "n", -1, NULL);
	} else {
		img = vips_image_new_from_buffer(buf, len, "", NULL);
	}

	if (!img) {
		libvips_error();
		return -1;
	}

	if (outType == 1) { // webp
		if (vips_webpsave_buffer(img, out_buf, out_size, "lossless", TRUE, NULL)) {
			g_object_unref(img);
			libvips_error();
			return -2;
		}
	} else if (outType == 2) { // png
		if (vips_pngsave_buffer(img, out_buf, out_size, NULL)) {
			g_object_unref(img);
			libvips_error();
			return -2;
		}
	} else if (outType == 3) { // jpeg
		if (vips_jpegsave_buffer(img, out_buf, out_size, NULL)) {
			g_object_unref(img);
			libvips_error();
			return -2;
		}
	} else if (outType == 4) { // gif
		if (vips_gifsave_buffer(img, out_buf, out_size, NULL)) {
			g_object_unref(img);
			libvips_error();
			return -2;
		}
	} else {
		g_object_unref(img);
		printf("libvips: unsupported out type: %d\n", outType);
		return -2;
	}

	g_object_unref(img);
	return 0;
}

void libvips_g_free(void* p) {
	g_free(p);
}
*/
import "C"
import (
	"log/slog"
	"unsafe"
)

type Format int

const (
	FORMAT_WEBP = Format(1)
	FORMAT_PNG  = Format(2)
	FORMAT_JEPG = Format(3)
	FORMAT_GIF  = Format(4)
)

func LibvipsInit() {
	if C.libvips_init() != 0 {
		panic("libvips init")
	}
	slog.Debug("libvips", "version", LibvipsVersion())
}

func LibvipsShutdown() {
	C.libvips_shutdown()
}

func LibvipsVersion() string {
	return C.GoString(C.libvips_version())
}

// encode in to target format
//
// return nil if any error occurred
func LibvipsEncode(in []byte, animated bool, target Format) []byte {
	cbytes := C.CBytes(in)

	var outBuf unsafe.Pointer
	var outSize C.size_t

	a := 0
	if animated {
		a = 1
	}

	result := C.libvips_encode((*C.char)(cbytes), C.int(len(in)), &outBuf, &outSize, C.int(int(target)), C.int(a))
	defer C.free(cbytes)
	if outBuf != nil {
		defer C.libvips_g_free(outBuf)
	}

	if result != 0 {
		return nil
	}

	buf := make([]byte, outSize)
	copy(buf, (*[1 << 30]byte)(outBuf)[:outSize:outSize])

	return buf
}
