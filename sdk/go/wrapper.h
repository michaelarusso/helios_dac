#ifndef HELIOS_WRAPPER_H
#define HELIOS_WRAPPER_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdbool.h>

// Opaque handle to the HeliosDac C++ class instance
typedef void* HeliosDacHandle;

// A simplified Point structure for the C API.
// Renamed to avoid name conflicts with C++ SDK types in wrapper.cpp.
// These structs MUST strictly match the memory layout of the C++ SDK structs.

typedef struct {
    uint16_t x;
    uint16_t y;
    uint8_t r;
    uint8_t g;
    uint8_t b;
    uint8_t i;
} WrapperHeliosPoint;

typedef struct {
    uint16_t x;
    uint16_t y;
    uint16_t r;
    uint16_t g;
    uint16_t b;
} WrapperHeliosPointHighRes;

typedef struct {
    uint16_t x;
    uint16_t y;
    uint16_t r;
    uint16_t g;
    uint16_t b;
    uint16_t i;
    uint16_t user1;
    uint16_t user2;
    uint16_t user3;
    uint16_t user4;
} WrapperHeliosPointExt;

// Constructor / Destructor
HeliosDacHandle HeliosDac_New();
void HeliosDac_Delete(HeliosDacHandle h);

// Device Management
int HeliosDac_OpenDevices(HeliosDacHandle h);
int HeliosDac_OpenDevicesOnlyUsb(HeliosDacHandle h);
int HeliosDac_OpenDevicesOnlyNetwork(HeliosDacHandle h);
void HeliosDac_CloseDevices(HeliosDacHandle h);
int HeliosDac_ReScanDevices(HeliosDacHandle h);
int HeliosDac_ReScanDevicesOnlyUsb(HeliosDacHandle h);
int HeliosDac_ReScanDevicesOnlyNetwork(HeliosDacHandle h);

// Device Info (using index 0 to numDevices-1)
// buffer must be at least 32 bytes
int HeliosDac_GetName(HeliosDacHandle h, int deviceIndex, char* buffer, int length);
// name must be max 20 chars (21 with null term)
int HeliosDac_SetName(HeliosDacHandle h, int deviceIndex, char* name);
bool HeliosDac_GetIsUsb(HeliosDacHandle h, int deviceIndex);
int HeliosDac_GetFirmwareVersion(HeliosDacHandle h, int deviceIndex);
int HeliosDac_GetSupportsHigherResolutions(HeliosDacHandle h, int deviceIndex);
bool HeliosDac_GetIsClosed(HeliosDacHandle h, int deviceIndex);
int HeliosDac_GetStatus(HeliosDacHandle h, int deviceIndex);

// Control
int HeliosDac_Stop(HeliosDacHandle h, int deviceIndex);
int HeliosDac_SetShutter(HeliosDacHandle h, int deviceIndex, bool level);
int HeliosDac_EraseFirmware(HeliosDacHandle h, int deviceIndex); // Advanced use only
int HeliosDac_SetLibusbDebugLogLevel(HeliosDacHandle h, int logLevel);

// Output
// pps: Points per second (e.g., 30000)
// flags: e.g. HELIOS_FLAGS_DEFAULT (value 0?)
int HeliosDac_WriteFrame(HeliosDacHandle h, int deviceIndex, int pps, int flags, const WrapperHeliosPoint* points, int numPoints);
int HeliosDac_WriteFrameHighResolution(HeliosDacHandle h, int deviceIndex, int pps, int flags, const WrapperHeliosPointHighRes* points, int numPoints);
int HeliosDac_WriteFrameExtended(HeliosDacHandle h, int deviceIndex, int pps, int flags, const WrapperHeliosPointExt* points, int numPoints);

#ifdef __cplusplus
}
#endif

#endif // HELIOS_WRAPPER_H
