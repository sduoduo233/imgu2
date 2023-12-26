package libvips

/*
#cgo pkg-config: vips libheif
#include <libheif/heif.h>
#include "vips/vips.h"
#include <stdio.h>
#include <stdlib.h>
#include <malloc.h>

void libvips_error() {
	printf("libvips: error: %s\n", vips_error_buffer());
	vips_error_clear();
}

int libvips_init() {
	if (VIPS_INIT("")) {
		libvips_error();
		return -1;
	}
	vips_leak_set(TRUE);
	vips_cache_set_max(0);
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
		img = vips_image_new_from_buffer(buf, len, "", "n", -1, "access", VIPS_ACCESS_SEQUENTIAL, NULL);
	} else {
		img = vips_image_new_from_buffer(buf, len, "", "access", VIPS_ACCESS_SEQUENTIAL, NULL);
	}

	if (!img) {
		libvips_error();
		return -1;
	}

	if (outType == 1) { // webp
		if (vips_webpsave_buffer(img, out_buf, out_size, "lossless", TRUE, NULL)) {
			g_object_unref(img);
			malloc_trim(0);
			libvips_error();
			return -2;
		}
	} else if (outType == 2) { // png
		if (vips_pngsave_buffer(img, out_buf, out_size, NULL)) {
			g_object_unref(img);
			malloc_trim(0);
			libvips_error();
			return -2;
		}
	} else if (outType == 3) { // jpeg
		if (vips_jpegsave_buffer(img, out_buf, out_size, NULL)) {
			g_object_unref(img);
			malloc_trim(0);
			libvips_error();
			return -2;
		}
	} else if (outType == 4) { // gif
		if (vips_gifsave_buffer(img, out_buf, out_size, NULL)) {
			g_object_unref(img);
			malloc_trim(0);
			libvips_error();
			return -2;
		}
	} else if (outType == 5) { // avif
		if (vips_heifsave_buffer(img, out_buf, out_size, "compression", VIPS_FOREIGN_HEIF_COMPRESSION_AV1, "encoder", VIPS_FOREIGN_HEIF_ENCODER_AOM, "lossless", TRUE, NULL)) {
			g_object_unref(img);
			malloc_trim(0);
			libvips_error();
			return -2;
		}
	} else {
		g_object_unref(img);
		malloc_trim(0);
		printf("libvips: unsupported out type: %d\n", outType);
		return -2;
	}

	g_object_unref(img);
	malloc_trim(0);
	return 0;
}

void libvips_g_free(void* p) {
	g_free(p);
	malloc_trim(0);
}

int libvips_heif_load_plugins(char* directory) {
	int nPlugins;
	struct heif_error error = heif_load_plugins(directory, NULL, &nPlugins, 0);
	if (error.code != heif_error_Ok) {
		printf("libvips: libvips_heif_load_plugins: %s\n", error.message);
		return -1;
	}
	return 0;
}
*/
import "C"
import (
	"log/slog"
	"os"
	"unsafe"
)

type Format int

const (
	FORMAT_WEBP = Format(1)
	FORMAT_PNG  = Format(2)
	FORMAT_JEPG = Format(3)
	FORMAT_GIF  = Format(4)
	FORMAT_AVIF = Format(5)
)

func LibvipsInit() {
	if C.libvips_init() != 0 {
		panic("libvips init")
	}
	slog.Debug("libvips", "version", LibvipsVersion())

	// For some unknown reason, libheif plugins are not loaded automatically
	// on Ubuntu (at least not on my Ubuntu 23.10). Loading these plugins
	// manually is required, or AVIF encoding will not work.

	// IMGU2_DEBUG_LIBHEIF_PLUGIN_PATHS is the path to the plugin directory.
	// For example, it might be set to /usr/lib/x86_64-linux-gnu/libheif/plugins
	// on Ubuntu 23.10
	libheifPlugin := os.Getenv("IMGU2_DEBUG_LIBHEIF_PLUGIN_PATHS")
	if libheifPlugin != "" {
		slog.Info("libvips_heif_load_plugins", "path", libheifPlugin)
		cstring := C.CString(libheifPlugin)
		defer C.free((unsafe.Pointer(cstring)))

		if C.libvips_heif_load_plugins(cstring) != 0 {
			panic("heif_load_plugins")
		}
	}
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
