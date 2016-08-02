package libvirt

/*
#cgo CFLAGS: -Wno-implicit-function-declaration
#cgo LDFLAGS: -lvirt
#include <libvirt/libvirt.h>
#include <libvirt/virterror.h>
#include <stdlib.h>

int domainEventLifecycleCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     int event, int detail, void *data)
{
    return domainEventLifecycleCallback(c, d, event, detail, data);
}

int domainEventGenericCallback_cgo(virConnectPtr c, virDomainPtr d, void *data)
{
    return domainEventGenericCallback(c, d, data);
}

int domainEventRTCChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                     long long utcoffset, void *data)
{
    return domainEventRTCChangeCallback(c, d, utcoffset, data);
}

int domainEventWatchdogCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int action, void *data)
{
    return domainEventWatchdogCallback(c, d, action, data);
}

int domainEventIOErrorCallback_cgo(virConnectPtr c, virDomainPtr d,
                                   const char *srcPath, const char *devAlias,
                                   int action, void *data)
{
    return domainEventIOErrorCallback(c, d, srcPath, devAlias, action, data);
}

int domainEventGraphicsCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    int phase, const virDomainEventGraphicsAddress *local,
                                    const virDomainEventGraphicsAddress *remote,
                                    const char *authScheme,
                                    const virDomainEventGraphicsSubject *subject, void *data)
{
    return domainEventGraphicsCallback(c, d, phase, local, remote, authScheme, subject, data);
}

int domainEventIOErrorReasonCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *srcPath, const char *devAlias,
                                         int action, const char *reason, void *data)
{
    return domainEventIOErrorReasonCallback(c, d, srcPath, devAlias, action, reason, data);
}

int domainEventBlockJobCallback_cgo(virConnectPtr c, virDomainPtr d,
                                    const char *disk, int type, int status, void *data)
{
    return domainEventIOErrorReasonCallback(c, d, disk, type, status, data);
}

int domainEventDiskChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *oldSrcPath, const char *newSrcPath,
                                      const char *devAlias, int reason, void *data)
{
    return domainEventDiskChangeCallback(c, d, oldSrcPath, newSrcPath, devAlias, reason, data);
}

int domainEventTrayChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                      const char *devAlias, int reason, void *data)
{
    return domainEventTrayChangeCallback(c, d, devAlias, reason, data);
}

int domainEventReasonCallback_cgo(virConnectPtr c, virDomainPtr d,
                                  int reason, void *data)
{
    return domainEventReasonCallback(c, d, reason, data);
}

int domainEventBalloonChangeCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         unsigned long long actual, void *data)
{
    return domainEventBalloonChangeCallback(c, d, actual, data);
}

int domainEventDeviceRemovedCallback_cgo(virConnectPtr c, virDomainPtr d,
                                         const char *devAlias, void *data)
{
    return domainEventDeviceRemovedCallback(c, d, devAlias, data);
}

void freeGoCallback_cgo(void* goCallbackId) {
   freeCallbackId((long)goCallbackId);
}

int virConnectDomainEventRegisterAny_cgo(virConnectPtr c,  virDomainPtr d,
                                         int eventID, virConnectDomainEventGenericCallback cb,
                                         long goCallbackId) {
    void* id = (void*)goCallbackId;
    return virConnectDomainEventRegisterAny(c, d, eventID, cb, id, freeGoCallback_cgo);
}

void errorGlobalCallback_cgo(void *userData, virErrorPtr error)
{
    globalErrorCallback(error);
}

void errorConnCallback_cgo(void *userData, virErrorPtr error)
{
    connErrorCallback((long)userData, error);
}

void virConnSetErrorFunc_cgo(virConnectPtr c, long goCallbackId, virErrorFunc cb)
{
    void* id = (void*)goCallbackId;
    virConnSetErrorFunc(c, id, cb);
}


*/
import "C"
