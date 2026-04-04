//go:build windows

package windows

// CGO bridge for IDWriteTextRenderer::DrawGlyphRun callback.
//
// On Windows x64, non-variadic COM calls place float parameters exclusively
// in XMM registers. Go's syscall.NewCallback cannot reliably extract floats
// from XMM when parameters are declared as uintptr, and declaring them as
// float32 causes deadlocks in reentrant COM callback chains.
//
// This C trampoline solves the problem: the C compiler correctly receives
// floats from XMM registers, then passes them to the Go //export function
// via CGO's mature bridging mechanism. The C code is statically linked into
// the Go binary — no external DLL is produced.

/*
#include <stdint.h>

// Forward declaration of the Go callback (defined below via //export).
extern uintptr_t goDrawGlyphRunBridge(
    uintptr_t thisPtr,
    uintptr_t clientCtx,
    float baselineOriginX,
    float baselineOriginY,
    uintptr_t measuringMode,
    uintptr_t glyphRun,
    uintptr_t glyphRunDescription,
    uintptr_t clientDrawingEffect
);

// C trampoline — receives the COM callback with proper float params.
// On Windows x64 there is only one calling convention, so no __stdcall needed.
static uintptr_t cDrawGlyphRunTrampoline(
    void* thisPtr,
    void* clientCtx,
    float baselineOriginX,
    float baselineOriginY,
    uint32_t measuringMode,
    void* glyphRun,
    void* glyphRunDescription,
    void* clientDrawingEffect)
{
    return goDrawGlyphRunBridge(
        (uintptr_t)thisPtr,
        (uintptr_t)clientCtx,
        baselineOriginX,
        baselineOriginY,
        (uintptr_t)measuringMode,
        (uintptr_t)glyphRun,
        (uintptr_t)glyphRunDescription,
        (uintptr_t)clientDrawingEffect
    );
}

// Returns the C function pointer for use in the COM vtable.
static uintptr_t getDrawGlyphRunTrampoline() {
    return (uintptr_t)cDrawGlyphRunTrampoline;
}
*/
import "C"

import (
	"math"
	"syscall"
	"unsafe"
)

//export goDrawGlyphRunBridge
func goDrawGlyphRunBridge(
	thisPtr C.uintptr_t,
	clientCtx C.uintptr_t,
	baselineOriginX C.float,
	baselineOriginY C.float,
	measuringMode C.uintptr_t,
	glyphRun C.uintptr_t,
	glyphRunDescription C.uintptr_t,
	clientDrawingEffect C.uintptr_t,
) C.uintptr_t {
	// thisPtr is a pointer to goTextRenderer (COM convention: object pointer
	// IS pointer-to-vtable-pointer, and vtable is our first struct field).
	tr := (*goTextRenderer)(unsafe.Pointer(uintptr(thisPtr)))
	tr.drawCallCount++
	if tr.bitmapTarget == 0 {
		return 0 // S_OK
	}

	// Convert float32 to uintptr via bit manipulation for SyscallN.
	xBits := uintptr(math.Float32bits(float32(baselineOriginX)))
	yBits := uintptr(math.Float32bits(float32(baselineOriginY)))

	// Call IDWriteBitmapRenderTarget::DrawGlyphRun (vtable index 3).
	vtablePtr := *(*uintptr)(unsafe.Pointer(tr.bitmapTarget))
	methodPtr := *(*uintptr)(unsafe.Pointer(vtablePtr + unsafe.Sizeof(uintptr(0))*3))
	var blackBoxRect RECT
	syscall.SyscallN(methodPtr,
		tr.bitmapTarget,
		xBits,
		yBits,
		uintptr(measuringMode),
		uintptr(glyphRun),
		tr.renderParams,
		uintptr(tr.textColor),
		uintptr(unsafe.Pointer(&blackBoxRect)),
	)
	return 0 // S_OK
}

// cgoDrawGlyphRunCallback returns the C trampoline function pointer
// for use in the IDWriteTextRenderer COM vtable.
func cgoDrawGlyphRunCallback() uintptr {
	return uintptr(C.getDrawGlyphRunTrampoline())
}
