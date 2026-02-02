#include "wrapper.h"
#include "sdk/cpp/HeliosDac.h"
#include <vector>
#include <cstring>
#include <algorithm>

extern "C" {

HeliosDacHandle HeliosDac_New() {
    return new HeliosDac();
}

void HeliosDac_Delete(HeliosDacHandle h) {
    if (h) {
        delete static_cast<HeliosDac*>(h);
    }
}

int HeliosDac_OpenDevices(HeliosDacHandle h) {
    return static_cast<HeliosDac*>(h)->OpenDevices();
}

int HeliosDac_OpenDevicesOnlyUsb(HeliosDacHandle h) {
    return static_cast<HeliosDac*>(h)->OpenDevicesOnlyUsb();
}

int HeliosDac_OpenDevicesOnlyNetwork(HeliosDacHandle h) {
    return static_cast<HeliosDac*>(h)->OpenDevicesOnlyNetwork();
}

void HeliosDac_CloseDevices(HeliosDacHandle h) {
    static_cast<HeliosDac*>(h)->CloseDevices();
}

int HeliosDac_ReScanDevices(HeliosDacHandle h) {
     return static_cast<HeliosDac*>(h)->ReScanDevices();
}

int HeliosDac_ReScanDevicesOnlyUsb(HeliosDacHandle h) {
     return static_cast<HeliosDac*>(h)->ReScanDevicesOnlyUsb();
}

int HeliosDac_ReScanDevicesOnlyNetwork(HeliosDacHandle h) {
     return static_cast<HeliosDac*>(h)->ReScanDevicesOnlyNetwork();
}

int HeliosDac_GetName(HeliosDacHandle h, int deviceIndex, char* buffer, int length) {
    return static_cast<HeliosDac*>(h)->GetName(deviceIndex, buffer);
}

int HeliosDac_SetName(HeliosDacHandle h, int deviceIndex, char* name) {
    return static_cast<HeliosDac*>(h)->SetName(deviceIndex, name);
}

bool HeliosDac_GetIsUsb(HeliosDacHandle h, int deviceIndex) {
    return static_cast<HeliosDac*>(h)->GetIsUsb(deviceIndex);
}

int HeliosDac_GetFirmwareVersion(HeliosDacHandle h, int deviceIndex) {
    return static_cast<HeliosDac*>(h)->GetFirmwareVersion(deviceIndex);
}

int HeliosDac_GetSupportsHigherResolutions(HeliosDacHandle h, int deviceIndex) {
    return static_cast<HeliosDac*>(h)->GetSupportsHigherResolutions(deviceIndex);
}

bool HeliosDac_GetIsClosed(HeliosDacHandle h, int deviceIndex) {
    return static_cast<HeliosDac*>(h)->GetIsClosed(deviceIndex);
}

int HeliosDac_GetStatus(HeliosDacHandle h, int deviceIndex) {
    return static_cast<HeliosDac*>(h)->GetStatus(deviceIndex);
}

int HeliosDac_Stop(HeliosDacHandle h, int deviceIndex) {
    return static_cast<HeliosDac*>(h)->Stop(deviceIndex);
}

int HeliosDac_SetShutter(HeliosDacHandle h, int deviceIndex, bool level) {
    return static_cast<HeliosDac*>(h)->SetShutter(deviceIndex, level);
}

int HeliosDac_EraseFirmware(HeliosDacHandle h, int deviceIndex) {
    return static_cast<HeliosDac*>(h)->EraseFirmware(deviceIndex);
}

int HeliosDac_SetLibusbDebugLogLevel(HeliosDacHandle h, int logLevel) {
    return static_cast<HeliosDac*>(h)->SetLibusbDebugLogLevel(logLevel);
}

int HeliosDac_WriteFrame(HeliosDacHandle h, int deviceIndex, int pps, int flags, const WrapperHeliosPoint* points, int numPoints) {
    if (!points || numPoints <= 0) return 0; // Or error code

    // Since the memory layout of WrapperHeliosPoint matches the SDK's HeliosPoint,
    // we can safely cast the pointer.
    return static_cast<HeliosDac*>(h)->WriteFrame(deviceIndex, pps, flags, (HeliosPoint*)points, numPoints);
}

int HeliosDac_WriteFrameHighResolution(HeliosDacHandle h, int deviceIndex, int pps, int flags, const WrapperHeliosPointHighRes* points, int numPoints) {
    if (!points || numPoints <= 0) return 0;
    return static_cast<HeliosDac*>(h)->WriteFrameHighResolution(deviceIndex, pps, flags, (HeliosPointHighRes*)points, numPoints);
}

int HeliosDac_WriteFrameExtended(HeliosDacHandle h, int deviceIndex, int pps, int flags, const WrapperHeliosPointExt* points, int numPoints) {
    if (!points || numPoints <= 0) return 0;
    return static_cast<HeliosDac*>(h)->WriteFrameExtended(deviceIndex, pps, flags, (HeliosPointExt*)points, numPoints);
}

}
